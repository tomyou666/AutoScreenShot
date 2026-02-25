//go:build windows

package ui

import (
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

const (
	BIF_RETURNONLYFSDIRS = 0x00000001
	BIF_NEWDIALOGSTYLE   = 0x00000040
)

var (
	shell32             = syscall.NewLazyDLL("shell32.dll")
	ole32               = syscall.NewLazyDLL("ole32.dll")
	shBrowseForFolder   = shell32.NewProc("SHBrowseForFolderW")
	shGetPathFromIDList = shell32.NewProc("SHGetPathFromIDListW")
	coTaskMemFree       = ole32.NewProc("CoTaskMemFree")
)

type BROWSEINFOW struct {
	Owner       win.HWND
	Root        uintptr
	DisplayName *uint16
	Title       *uint16
	Flags       uint32
	Callback    uintptr
	LParam      uintptr
	Image       int32
}

func browseForFolder(owner walk.Form) (string, error) {
	_ = owner
	title, _ := syscall.UTF16PtrFromString("保存先フォルダを選択")
	bi := BROWSEINFOW{
		Title: title,
		Flags: BIF_RETURNONLYFSDIRS | BIF_NEWDIALOGSTYLE,
	}
	pidl, _, _ := shBrowseForFolder.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return "", nil
	}
	defer coTaskMemFree.Call(pidl)

	buf := make([]uint16, win.MAX_PATH)
	shGetPathFromIDList.Call(pidl, uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf), nil
}
