package output

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
)

const defaultJpegQuality = 85

// SaveJPG は画像を指定フォルダに連番の JPG として保存し、ファイルパスを返します。
func SaveJPG(dir string, index int, img image.Image, quality int) (string, error) {
	if quality <= 0 {
		quality = defaultJpegQuality
	}
	name := fmt.Sprintf("screenshot_%05d.jpg", index)
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: quality}); err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}
