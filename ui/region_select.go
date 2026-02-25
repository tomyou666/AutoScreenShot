//go:build windows

package ui

import (
	"sync"
	"syscall"
	"unsafe"

	"github.com/kbinani/screenshot"
	"github.com/lxn/win"
)

// Region はキャプチャ範囲を表します。
type Region struct {
	X, Y, Width, Height int
}

// SelectRegion は全画面オーバーレイを表示し、マウスドラッグで矩形を選択させます。
// 選択された範囲と true を返します。Esc でキャンセルした場合は false を返します。
func SelectRegion() (Region, bool) {
	bx, by, bw, bh := virtualScreenBounds()
	if bw <= 0 || bh <= 0 {
		return Region{}, false
	}

	var result Region
	var ok bool
	var mu sync.Mutex

	var startX, startY int
	var currentX, currentY int
	dragging := false

	const wndClassName = "AutoScreenShotRegionSelect"

	win.RegisterClassEx(&win.WNDCLASSEX{
		CbSize: uint32(unsafe.Sizeof(win.WNDCLASSEX{})),
		Style:  win.CS_HREDRAW | win.CS_VREDRAW,
		LpfnWndProc: syscall.NewCallback(func(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
			switch msg {
			case win.WM_LBUTTONDOWN:
				mu.Lock()
				startX = int(win.LOWORD(uint32(lParam)))
				startY = int(win.HIWORD(uint32(lParam)))
				currentX = startX
				currentY = startY
				dragging = true
				mu.Unlock()
				win.InvalidateRect(hwnd, nil, true)
				return 0
			case win.WM_MOUSEMOVE:
				if wParam&win.MK_LBUTTON != 0 {
					mu.Lock()
					currentX = int(win.LOWORD(uint32(lParam)))
					currentY = int(win.HIWORD(uint32(lParam)))
					mu.Unlock()
					win.InvalidateRect(hwnd, nil, true)
				}
				return 0
			case win.WM_LBUTTONUP:
				mu.Lock()
				if dragging {
					ex := int(win.LOWORD(uint32(lParam)))
					ey := int(win.HIWORD(uint32(lParam)))
					x1, x2 := startX, ex
					if x1 > x2 {
						x1, x2 = x2, x1
					}
					y1, y2 := startY, ey
					if y1 > y2 {
						y1, y2 = y2, y1
					}
					w := x2 - x1
					h := y2 - y1
					if w > 2 && h > 2 {
						result = Region{X: x1 + bx, Y: y1 + by, Width: w, Height: h}
						ok = true
					}
					dragging = false
					win.PostQuitMessage(0)
				}
				mu.Unlock()
				return 0
			case win.WM_KEYDOWN:
				if wParam == win.VK_ESCAPE {
					win.PostQuitMessage(0)
					return 0
				}
				return 0
			case win.WM_PAINT:
				var ps win.PAINTSTRUCT
				hdc := win.BeginPaint(hwnd, &ps)
				if hdc != 0 {
					mu.Lock()
					sx, sy := startX, startY
					cx, cy := currentX, currentY
					dr := dragging
					mu.Unlock()
					if dr {
						x1, x2 := sx, cx
						if x1 > x2 {
							x1, x2 = x2, x1
						}
						y1, y2 := sy, cy
						if y1 > y2 {
							y1, y2 = y2, y1
						}
						pen := createPen(win.PS_SOLID, 3, uint32(win.RGB(255, 0, 0)))
						oldPen := win.SelectObject(hdc, win.HGDIOBJ(pen))
						win.SelectObject(hdc, win.GetStockObject(win.NULL_BRUSH))
						win.Rectangle_(hdc, int32(x1), int32(y1), int32(x2), int32(y2))
						win.SelectObject(hdc, oldPen)
						win.DeleteObject(win.HGDIOBJ(pen))
					}
				}
				win.EndPaint(hwnd, &ps)
				return 0
			case win.WM_DESTROY:
				return 0
			}
			return win.DefWindowProc(hwnd, msg, wParam, lParam)
		}),
		HInstance:     win.GetModuleHandle(nil),
		LpszClassName: syscall.StringToUTF16Ptr(wndClassName),
		HCursor:       win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_CROSS)),
		HbrBackground: win.HBRUSH(win.COLOR_WINDOW + 1),
	})

	hwnd := win.CreateWindowEx(
		win.WS_EX_LAYERED|win.WS_EX_TOPMOST|win.WS_EX_TOOLWINDOW,
		syscall.StringToUTF16Ptr(wndClassName),
		nil,
		win.WS_POPUP|win.WS_VISIBLE,
		int32(bx), int32(by), int32(bw), int32(bh),
		0, 0, win.GetModuleHandle(nil), nil,
	)
	if hwnd == 0 {
		return Region{}, false
	}
	setLayeredWindowAttributes(hwnd, 0, 180, 0x2) // LWA_ALPHA = 0x2

	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) != 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}

	win.DestroyWindow(hwnd)

	return result, ok
}

func virtualScreenBounds() (x, y, width, height int) {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return 0, 0, 1920, 1080
	}
	minX, minY := 0, 0
	maxX, maxY := 0, 0
	for i := 0; i < n; i++ {
		b := screenshot.GetDisplayBounds(i)
		if b.Min.X < minX {
			minX = b.Min.X
		}
		if b.Min.Y < minY {
			minY = b.Min.Y
		}
		if b.Max.X > maxX {
			maxX = b.Max.X
		}
		if b.Max.Y > maxY {
			maxY = b.Max.Y
		}
	}
	return minX, minY, maxX - minX, maxY - minY
}

var (
	gdi32CreatePen   = syscall.NewLazyDLL("gdi32.dll").NewProc("CreatePen")
	user32SetLayered = syscall.NewLazyDLL("user32.dll").NewProc("SetLayeredWindowAttributes")
)

func createPen(style, width int32, color uint32) win.HPEN {
	r, _, _ := gdi32CreatePen.Call(uintptr(style), uintptr(width), uintptr(color))
	return win.HPEN(r)
}

func setLayeredWindowAttributes(hwnd win.HWND, crKey uint32, bAlpha uint8, dwFlags uint32) bool {
	r, _, _ := user32SetLayered.Call(uintptr(hwnd), uintptr(crKey), uintptr(bAlpha), uintptr(dwFlags))
	return r != 0
}
