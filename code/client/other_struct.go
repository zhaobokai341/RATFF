package main

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type FileCopy struct {
	source string
	target string
}

type FileCompress struct {
	source string
	target string
}

// 复制文件
func (fcpy *FileCopy) copy_file(source, target string) error {
	old_file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer old_file.Close()
	new_file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer new_file.Close()
	_, err = io.Copy(new_file, old_file)
	if err != nil {
		return err
	}
	return nil
}

// 递归复制文件夹
func (fcpy *FileCopy) copy_dir(source, target string) error {
	err_info := ""
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(target, sourceInfo.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(source, entry.Name())
		dstPath := filepath.Join(target, entry.Name())

		if entry.IsDir() {
			err = fcpy.copy_dir(srcPath, dstPath)
			if err != nil {
				err_info += err.Error() + "\n"
			}
		} else {
			err = fcpy.copy_file(srcPath, dstPath)
			if err != nil {
				err_info += err.Error() + "\n"
			}
		}
	}

	if err_info != "" {
		return errors.New(err_info)
	}

	return nil
}

// 压缩文件
func (fcmps *FileCompress) compress_file(source, target string) error {
	source_file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer source_file.Close()
	new_file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer new_file.Close()
	writer := zip.NewWriter(new_file)
	defer writer.Close()
	file, err := writer.Create(filepath.Base(source))
	if err != nil {
		return err
	}
	_, err = io.Copy(file, source_file)
	if err != nil {
		return err
	}
	return nil
}

// 压缩文件夹
func (fcmps *FileCompress) compress_dir(source, target string) error {
	err_info := ""
	zip_file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zip_file.Close()

	writer := zip.NewWriter(zip_file)
	defer writer.Close()

	return filepath.Walk(source, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			err_info += err.Error() + "\n"
		}

		relPath, err := filepath.Rel(source, filePath)
		if err != nil {
			err_info += err.Error() + "\n"
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			err_info += err.Error() + "\n"
		}

		header.Name = relPath
		if info.IsDir() {
			header.Name += "/"
		} else {
			writer, err := writer.CreateHeader(header)
			if err != nil {
				err_info += err.Error() + "\n"
			}

			file, err := os.Open(filePath)
			if err != nil {
				err_info += err.Error() + "\n"
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err != nil {
				err_info += err.Error() + "\n"
			}
		}
		if err_info != "" {
			return errors.New(err_info)
		}
		return nil
	})
}
