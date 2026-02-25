//go:build windows

package focus

import (
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

var (
	user32             = syscall.NewLazyDLL("user32.dll")
	procEnumWindows    = user32.NewProc("EnumWindows")
	procGetWindowTextW = user32.NewProc("GetWindowTextW")
)

// ListVisibleWindowTitles は表示されているトップレベルウィンドウのタイトル一覧を返します。
func ListVisibleWindowTitles() []string {
	var titles []string
	cb := syscall.NewCallback(func(hwnd win.HWND, lParam uintptr) uintptr {
		if win.IsWindowVisible(hwnd) {
			buf := make([]uint16, 256)
			n, _ := getWindowText(hwnd, buf)
			if n > 0 {
				title := syscall.UTF16ToString(buf[:n])
				titlesPtr := (*[]string)(unsafe.Pointer(lParam))
				*titlesPtr = append(*titlesPtr, title)
			}
		}
		return 1 // 続行
	})
	_, _, _ = procEnumWindows.Call(cb, uintptr(unsafe.Pointer(&titles)))
	return titles
}

func getWindowText(hwnd win.HWND, buf []uint16) (int, error) {
	r0, _, e1 := syscall.Syscall(procGetWindowTextW.Addr(), 3,
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)))
	n := int(r0)
	if n == 0 && e1 != 0 {
		return 0, e1
	}
	return n, nil
}

// SetForegroundByTitle は指定したタイトルに完全一致する最初の表示中ウィンドウを前面にします。
// 見つからなければ false を返します。
func SetForegroundByTitle(title string) bool {
	var found win.HWND
	cb := syscall.NewCallback(func(hwnd win.HWND, lParam uintptr) uintptr {
		if win.IsWindowVisible(hwnd) {
			buf := make([]uint16, 256)
			n, _ := getWindowText(hwnd, buf)
			if n > 0 && syscall.UTF16ToString(buf[:n]) == title {
				found = hwnd
				return 0 // 列挙中止
			}
		}
		return 1
	})
	_, _, _ = procEnumWindows.Call(cb, 0)
	if found == 0 {
		return false
	}
	return win.SetForegroundWindow(found)
}
