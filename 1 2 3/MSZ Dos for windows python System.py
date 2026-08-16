import time
import platform
import ctypes
from ctypes import wintypes
import base64
import tempfile
import os

SystemStorage_One = "A:\\"
SystemStorage_Two = "B:\\"
CurrentSystemPath = SystemStorage_One
audio_alias = "audio"

def check_virt_registry():
    advapi32 = ctypes.WinDLL("advapi32", use_last_error=True)
    hkey = wintypes.HKEY()
    reg_path = r"HARDWARE\DESCRIPTION\System\CentralProcessor\0"
    HKEY_LOCAL_MACHINE = wintypes.HKEY(0x80000002)
    KEY_READ = 0x20019

    ret = advapi32.RegOpenKeyExW(
        HKEY_LOCAL_MACHINE,
        ctypes.c_wchar_p(reg_path),
        0,
        KEY_READ,
        ctypes.byref(hkey)
    )
    if ret != 0:
        return False

    data = wintypes.DWORD()
    data_len = wintypes.DWORD(ctypes.sizeof(data))
    advapi32.RegQueryValueExW(
        hkey,
        ctypes.c_wchar_p("FeatureSet"),
        None,
        None,
        ctypes.byref(data),
        ctypes.byref(data_len)
    )
    advapi32.RegCloseKey(hkey)
    virt_enable = (data.value >> 13) & 1
    return virt_enable == 1

virt_status = "Yes" if check_virt_registry() else "No"
virtualization = f"virtualization:{virt_status}"

winmm = ctypes.WinDLL("winmm.dll")
def mci(s):
    buf = ctypes.create_unicode_buffer(256)
    winmm.mciSendStringW(s, buf, 256, None)

while True:
    system = input("<" + CurrentSystemPath + ">")
    if system == "cd B:":
        CurrentSystemPath = SystemStorage_Two
    elif system == "cd A:":
        CurrentSystemPath = SystemStorage_One
    elif system == "system about":
        print("Operating System:", platform.system())
        print("System Version:", platform.version())
        print("CPU Name:", platform.processor())
        print("Full Platform Info:", platform.platform())
        print("Machine Architecture:", platform.machine())
        print(virtualization)
        print("------------system about------------")
        print("-----MSZ Dos version for 1.0.0------")
    elif system == "music list":
        print("------------music list-----------------")
        print("Bad apple")
    elif system == "Bad apple":
        mci(f"close {audio_alias}")
        with open("music1.txt", "r", encoding="utf-8") as f:
            b64_str = f.read()
        chunk = 4096
        audio_bin = b""
        for i in range(0, len(b64_str), chunk):
            audio_bin += base64.b64decode(b64_str[i:i+chunk])
        tmp_file = tempfile.NamedTemporaryFile(suffix=".mp3", delete=False)
        tmp_file.write(audio_bin)
        tmp_file.close()
        mci(f'open "{tmp_file.name}" type mpegvideo alias {audio_alias}')
        mci(f"play {audio_alias}")
        os.remove(tmp_file.name)
    elif system == "stop":
        mci(f"close {audio_alias}")
        print("Music stopped")
    elif system == "exit":
        break
    else:
        print(f"Unknown command: {system}")
