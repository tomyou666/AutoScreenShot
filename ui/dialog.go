//go:build windows

package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"AutoScreenShot/focus"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

// Settings はメインループに渡す設定です。
type Settings struct {
	Region           Region
	OutputFolder     string
	KeyOperation     string
	FocusWindowTitle string // 開始前にフォーカスするウィンドウのタイトル（空なら行わない）
	MaxCount         int
	StopOnThreeSame  bool
	DelayMsAfterKey  int
	PDFTitle         string // PDFのタイトル（デフォルトは screenshot-YYYY-MM-DD_HH-MM-SS）
}

// RunSettingsDialog は設定ダイアログを表示し、ユーザーが「開始」を押したとき設定を返します。
// キャンセル時は ok が false です。
func RunSettingsDialog() (Settings, bool) {
	var dlg *walk.Dialog
	var folderEdit, keyEdit, pdfTitleEdit *walk.LineEdit
	var focusCombo *walk.ComboBox
	var maxCountEdit *walk.NumberEdit
	var delayEdit *walk.NumberEdit
	var stopThreeCheck *walk.CheckBox
	var regionLabel *walk.Label
	var startBtn *walk.PushButton

	settings := Settings{
		KeyOperation:    "Enter",
		MaxCount:        500,
		StopOnThreeSame: true,
		DelayMsAfterKey: 500,
		PDFTitle:        "screenshot-" + time.Now().Format("2006-01-02_15-04-05"),
	}

	dlg, err := walk.NewDialog(nil)
	if err != nil {
		showError(fmt.Sprintf("ダイアログの作成に失敗しました: %v", err))
		return settings, false
	}
	dlg.SetTitle("自動スクリーンショット - 設定")
	dlg.SetLayout(walk.NewVBoxLayout())

	// 範囲選択
	regionComp, err := walk.NewComposite(dlg)
	if err != nil {
		showError(fmt.Sprintf("UIの作成に失敗しました: %v", err))
		dlg.Dispose()
		return settings, false
	}
	regionComp.SetLayout(walk.NewHBoxLayout())
	if l, err := walk.NewLabel(regionComp); err == nil {
		l.SetText("範囲:")
	}
	regionLabel, _ = walk.NewLabel(regionComp)
	regionLabel.SetText("(未選択)")
	selectRegionBtn, _ := walk.NewPushButton(regionComp)
	selectRegionBtn.SetText("範囲を選択...")
	selectRegionBtn.Clicked().Attach(func() {
		reg, ok := SelectRegion()
		if ok {
			settings.Region = reg
			regionLabel.SetText(strconv.Itoa(reg.Width) + " x " + strconv.Itoa(reg.Height))
		}
	})

	// 保存フォルダ
	folderComp, _ := walk.NewComposite(dlg)
	folderComp.SetLayout(walk.NewHBoxLayout())
	if l, err := walk.NewLabel(folderComp); err == nil {
		l.SetText("保存先:")
	}
	folderEdit, _ = walk.NewLineEdit(folderComp)
	folderEdit.SetText(settings.OutputFolder)
	browseBtn, _ := walk.NewPushButton(folderComp)
	browseBtn.SetText("参照...")
	browseBtn.Clicked().Attach(func() {
		path, err := runFolderBrowse(dlg)
		if err == nil && path != "" {
			if !isDirEmpty(path) {
				if showConfirm("確認", "選択したフォルダは空ではありません。フォルダを空にしますか？") {
					if err := emptyDir(path); err != nil {
						showError(fmt.Sprintf("フォルダを空にできませんでした: %v", err))
						return
					}
				}
			}
			folderEdit.SetText(path)
		}
	})

	// キー操作（フォーカス時にキーを押すとそのキーで設定される）
	keyComp, _ := walk.NewComposite(dlg)
	keyComp.SetLayout(walk.NewHBoxLayout())
	if l, err := walk.NewLabel(keyComp); err == nil {
		l.SetText("キー操作:")
	}
	keyEdit, _ = walk.NewLineEdit(keyComp)
	keyEdit.SetText(settings.KeyOperation)
	keyEdit.SetReadOnly(true)
	keyEdit.SetToolTipText("この欄をクリックしてフォーカスを移し、設定したいキーを押してください")
	keyEdit.KeyDown().Attach(func(key walk.Key) {
		// Enter は「開始」ボタンのトリガーに使うため、キー設定ではキャッチしない
		if key == walk.KeyReturn {
			return
		}
		s := keyOperationString(walk.ModifiersDown(), key)
		if s != "" {
			keyEdit.SetText(s)
		}
	})
	if l, err := walk.NewLabel(keyComp); err == nil {
		l.SetText(" (欄をクリックしてキーを押すと設定)")
	}

	// フォーカスするアプリケーション
	focusComp, _ := walk.NewComposite(dlg)
	focusComp.SetLayout(walk.NewHBoxLayout())
	if l, err := walk.NewLabel(focusComp); err == nil {
		l.SetText("フォーカスするアプリ:")
	}
	focusCombo, _ = walk.NewComboBox(focusComp)
	refreshFocusList := func() {
		titles := focus.ListVisibleWindowTitles()
		items := make([]string, 0, len(titles)+1)
		items = append(items, "(なし)")
		items = append(items, titles...)
		focusCombo.SetModel(items)
		if settings.FocusWindowTitle == "" {
			focusCombo.SetCurrentIndex(0)
		} else {
			for i, t := range items {
				if t == settings.FocusWindowTitle {
					focusCombo.SetCurrentIndex(i)
					break
				}
			}
		}
	}
	refreshFocusList()
	refreshFocusBtn, _ := walk.NewPushButton(focusComp)
	refreshFocusBtn.SetText("一覧を更新")
	refreshFocusBtn.Clicked().Attach(refreshFocusList)

	// 終了条件
	endComp, _ := walk.NewComposite(dlg)
	endComp.SetLayout(walk.NewHBoxLayout())
	if l, err := walk.NewLabel(endComp); err == nil {
		l.SetText("最大枚数 (0=無制限):")
	}
	maxCountEdit, _ = walk.NewNumberEdit(endComp)
	maxCountEdit.SetRange(0, 99999)
	maxCountEdit.SetValue(float64(settings.MaxCount))
	stopThreeCheck, _ = walk.NewCheckBox(endComp)
	stopThreeCheck.SetText("3枚連続同一で終了")
	stopThreeCheck.SetChecked(settings.StopOnThreeSame)

	// 待機時間
	delayComp, _ := walk.NewComposite(dlg)
	delayComp.SetLayout(walk.NewHBoxLayout())
	if l, err := walk.NewLabel(delayComp); err == nil {
		l.SetText("キー送信後の待機(ms):")
	}
	delayEdit, _ = walk.NewNumberEdit(delayComp)
	delayEdit.SetRange(0, 10000)
	delayEdit.SetValue(float64(settings.DelayMsAfterKey))

	// PDFタイトル
	pdfTitleComp, _ := walk.NewComposite(dlg)
	pdfTitleComp.SetLayout(walk.NewHBoxLayout())
	if l, err := walk.NewLabel(pdfTitleComp); err == nil {
		l.SetText("PDFタイトル:")
	}
	pdfTitleEdit, _ = walk.NewLineEdit(pdfTitleComp)
	pdfTitleEdit.SetText(settings.PDFTitle)
	pdfTitleEdit.SetToolTipText("PDFのメタデータタイトル。")

	// ボタン
	btnComp, _ := walk.NewComposite(dlg)
	btnComp.SetLayout(walk.NewHBoxLayout())
	_, _ = walk.NewHSpacer(btnComp)
	startBtn, _ = walk.NewPushButton(btnComp)
	startBtn.SetText("開始")
	startBtn.Clicked().Attach(func() {
		settings.OutputFolder = folderEdit.Text()
		settings.KeyOperation = keyEdit.Text()
		if t := focusCombo.Text(); t == "(なし)" || t == "" {
			settings.FocusWindowTitle = ""
		} else {
			settings.FocusWindowTitle = t
		}
		settings.MaxCount = int(maxCountEdit.Value())
		settings.StopOnThreeSame = stopThreeCheck.Checked()
		settings.DelayMsAfterKey = int(delayEdit.Value())
		if s := pdfTitleEdit.Text(); s != "" {
			settings.PDFTitle = s
		} else {
			settings.PDFTitle = "screenshot-" + time.Now().Format("2006-01-02_15-04-05")
		}
		// 必須項目のチェック（ダイアログを閉じる前に表示する）
		if settings.Region.Width <= 0 || settings.Region.Height <= 0 {
			showError("キャプチャ範囲を選択してください。「範囲を選択...」で範囲を指定してください。")
			return
		}
		if settings.OutputFolder == "" {
			showError("保存先フォルダを指定してください。")
			return
		}
		dlg.Accept()
	})
	cancelBtn, _ := walk.NewPushButton(btnComp)
	cancelBtn.SetText("キャンセル")
	cancelBtn.Clicked().Attach(func() {
		dlg.Cancel()
	})

	dlg.SetDefaultButton(startBtn)
	dlg.SetCancelButton(cancelBtn)

	if dlg.Run() != walk.DlgCmdOK {
		return settings, false
	}

	if settings.Region.Width <= 0 || settings.Region.Height <= 0 {
		return settings, false
	}
	if settings.OutputFolder == "" {
		return settings, false
	}

	return settings, true
}

// showError はエラーメッセージをメッセージボックスで表示します。
func showError(msg string) {
	title, _ := syscall.UTF16PtrFromString("エラー")
	text, _ := syscall.UTF16PtrFromString(msg)
	win.MessageBox(0, text, title, win.MB_OK|win.MB_ICONERROR)
}

// ShowInfo は情報メッセージをメッセージボックスで表示します。
func ShowInfo(title, msg string) {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(msg)
	win.MessageBox(0, m, t, win.MB_OK|win.MB_ICONINFORMATION)
}

// showConfirm は「はい」「いいえ」の確認メッセージを表示し、「はい」なら true を返します。
func showConfirm(title, msg string) bool {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(msg)
	return win.MessageBox(0, m, t, win.MB_YESNO|win.MB_ICONQUESTION) == win.IDYES
}

// isDirEmpty は指定フォルダが空（ファイル・サブフォルダが無い）場合に true を返します。
func isDirEmpty(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return true // 読めない場合は空とみなす
	}
	return len(entries) == 0
}

// emptyDir は指定フォルダ内のすべてのファイルとサブフォルダを削除します。フォルダ自体は削除しません。
func emptyDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if e.IsDir() {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
		} else {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}
	return nil
}

// runFolderBrowse は Windows のフォルダ選択ダイアログを表示します。
func runFolderBrowse(owner walk.Form) (string, error) {
	return browseForFolder(owner)
}
