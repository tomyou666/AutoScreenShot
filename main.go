//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"AutoScreenShot/capture"
	"AutoScreenShot/compare"
	"AutoScreenShot/focus"
	"AutoScreenShot/keyboard"
	"AutoScreenShot/output"
	"AutoScreenShot/ui"
)

func main() {
	// Windows GUI はメインスレッドで実行する必要がある
	runtime.LockOSThread()

	settings, ok := ui.RunSettingsDialog()
	if !ok {
		os.Exit(0)
		return
	}

	dir := settings.OutputFolder
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "フォルダ作成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// フォーカスするアプリが指定されていれば、そのウィンドウを前面にする
	if settings.FocusWindowTitle != "" {
		if focus.SetForegroundByTitle(settings.FocusWindowTitle) {
			time.Sleep(300 * time.Millisecond) // ウィンドウが前面になるまで待つ
		}
	}

	region := capture.Region{
		X:      settings.Region.X,
		Y:      settings.Region.Y,
		Width:  settings.Region.Width,
		Height: settings.Region.Height,
	}

	var prevHash, prevPrevHash []byte
	count := 0
	stoppedByThreeSame := false
	delay := time.Duration(settings.DelayMsAfterKey) * time.Millisecond
	if delay <= 0 {
		delay = 500 * time.Millisecond
	}

	for {
		img, err := capture.Capture(region)
		if err != nil {
			fmt.Fprintf(os.Stderr, "キャプチャに失敗しました: %v\n", err)
			break
		}

		count++
		path, err := output.SaveJPG(dir, count, img, 85)
		if err != nil {
			fmt.Fprintf(os.Stderr, "保存に失敗しました: %v\n", err)
			break
		}
		_ = path

		hash, err := compare.Hash(img, 85)
		if err != nil {
			break
		}

		if settings.MaxCount > 0 && count >= settings.MaxCount {
			break
		}
		if settings.StopOnThreeSame && compare.ThreeSame(prevPrevHash, prevHash, hash) {
			stoppedByThreeSame = true
			break
		}

		prevPrevHash = prevHash
		prevHash = hash

		if err := keyboard.Send(settings.KeyOperation); err != nil {
			fmt.Fprintf(os.Stderr, "キー送信に失敗しました: %v\n", err)
		}
		time.Sleep(delay)
	}

	// 3枚連続同一で終了した場合、同一の3枚のうち最後の2枚を削除してからPDF化する
	if stoppedByThreeSame && count >= 3 {
		for _, n := range []int{count, count - 1} {
			p := filepath.Join(dir, fmt.Sprintf("screenshot_%05d.jpg", n))
			if err := os.Remove(p); err != nil {
				fmt.Fprintf(os.Stderr, "重複画像の削除に失敗しました %s: %v\n", p, err)
			}
		}
	}

	pdfFileName := sanitizePDFFileName(settings.PDFTitle)
	if pdfFileName == "" {
		pdfFileName = "screenshots.pdf"
	} else if !strings.HasSuffix(strings.ToLower(pdfFileName), ".pdf") {
		pdfFileName += ".pdf"
	}
	pdfPath := filepath.Join(dir, pdfFileName)
	if err := output.JPGsToPDF(dir, pdfPath, settings.PDFTitle, settings.Region.Width, settings.Region.Height); err != nil {
		fmt.Fprintf(os.Stderr, "PDF生成に失敗しました: %v\n", err)
		os.Exit(1)
	}
	if stoppedByThreeSame && count >= 3 {
		fmt.Printf("完了: %d 枚保存（同一3枚のうち2枚を削除）、%s に PDF を出力しました。\n", count, pdfPath)
	} else {
		fmt.Printf("完了: %d 枚のスクリーンショットを保存し、%s に PDF を出力しました。\n", count, pdfPath)
	}
	ui.ShowInfo("完了", "完了しました。")
}

// sanitizePDFFileName はPDFタイトルをWindowsのファイル名として使えるように無効文字を除去します。
func sanitizePDFFileName(title string) string {
	const invalid = `\/:*?"<>|`
	s := strings.TrimSpace(title)
	var b strings.Builder
	for _, r := range s {
		if !strings.ContainsRune(invalid, r) && r >= 0x20 {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
