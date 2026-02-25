package output

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/jung-kurt/gofpdf"
)

// PixelsToMm は 96 DPI を基準にピクセルを mm に変換します。
const pixelsPerInch = 96
const mmPerInch = 25.4

func pixelsToMm(pixels int) float64 {
	return float64(pixels) * mmPerInch / pixelsPerInch
}

// JPGsToPDF は指定フォルダ内の JPG をファイル名順で1つの PDF に結合し、outPath に保存します。
// widthPx, heightPx はダイアログで設定したキャプチャ範囲（ピクセル）で、PDF のページサイズに反映されます。
// title はPDFのメタデータタイトルです。
func JPGsToPDF(dir, outPath, title string, widthPx, heightPx int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var jpgs []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".jpg" && ext != ".jpeg" {
			continue
		}
		jpgs = append(jpgs, filepath.Join(dir, e.Name()))
	}
	if len(jpgs) == 0 {
		return nil
	}
	sort.Strings(jpgs)

	// ダイアログで設定した範囲を PDF のページサイズ（mm）に変換
	wMm := pixelsToMm(widthPx)
	hMm := pixelsToMm(heightPx)
	if wMm <= 0 || hMm <= 0 {
		wMm, hMm = 210, 297 // フォールバック: A4
	}

	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		SizeStr:        "",
		FontDirStr:     "",
		Size:           gofpdf.SizeType{Wd: wMm, Ht: hMm},
	})
	if title != "" {
		pdf.SetTitle(title, true) // true = UTF-8（日本語対応）
	}
	for _, path := range jpgs {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.Size() == 0 {
			continue
		}
		opt := gofpdf.ImageOptions{ImageType: "JPEG"}
		pdf.AddPage()
		w, h := pdf.GetPageSize()
		pdf.ImageOptions(path, 0, 0, w, h, false, opt, 0, "")
	}
	return pdf.OutputFileAndClose(outPath)
}
