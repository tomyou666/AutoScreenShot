//go:build windows

package capture

import (
	"image"

	"github.com/kbinani/screenshot"
)

// Region はキャプチャ範囲（左上座標と幅・高さ）を表します。
type Region struct {
	X, Y, Width, Height int
}

// Capture は指定範囲をキャプチャして image.Image を返します。
func Capture(region Region) (image.Image, error) {
	bounds := image.Rect(region.X, region.Y, region.X+region.Width, region.Y+region.Height)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, err
	}
	return img, nil
}
