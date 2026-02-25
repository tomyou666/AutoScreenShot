# AutoScreenShot

Windows 向けの自動スクリーンショットアプリ（Go 製）です。  
指定範囲をキャプチャし、キー操作ごとに JPG を保存し、終了後に PDF にまとめます。

## 必要な環境

- Windows
- Go 1.21 以上

## ビルド

**重要:** walk を使うため、Common Controls v6 を要求する**アプリケーション マニフェスト**を exe に埋め込む必要があります（埋め込まないと「TTM_ADDTOOL failed」でダイアログが開きません）。

### 方法 A: マニフェストを exe に埋め込む（推奨）

1. [rsrc](https://github.com/akavel/rsrc) をインストール:
   ```bash
   go install github.com/akavel/rsrc@latest
   ```
2. マニフェストからリソースを生成:
   ```bash
   rsrc -manifest app.manifest -o rsrc.syso
   ```
3. ビルド（`rsrc.syso` が同じフォルダにあれば自動でリンクされます）:
   ```bash
   go build -ldflags="-H windowsgui" -o AutoScreenShot.exe
   ```

### 方法 B: マニフェストを exe と同じフォルダに置く

1. `go build -ldflags="-H windowsgui" -o AutoScreenShot.exe` でビルド
2. ビルドした exe と同じフォルダに `app.manifest` をコピーし、**exe の名前 + `.manifest`** にリネームする  
   例: `AutoScreenShot.exe` なら `AutoScreenShot.exe.manifest`
3. 中身はそのままでよい（`app.manifest` をリネームして使う）

## リリース

タグを push すると Windows 用 exe がビルドされ、GitHub の Releases に自動で公開されます。

1. リリース用のタグを打つ（例: `v1.0.0`）:
   ```bash
   git tag v1.0.0
   ```
2. タグを push する:
   ```bash
   git push origin v1.0.0
   ```
3. GitHub の Actionsでワークフローが走り、完了後に [Releases](https://github.com/tomyou666/AutoScreenShot/releases) に以下がアップロードされます:
   - `AutoScreenShot-windows-amd64.exe`（64bit Intel/AMD）
   - `AutoScreenShot-windows-arm64.exe`（64bit ARM）

既存のタグを付け直したい場合は `git tag -d v1.0.0` で削除してから再度タグを打ち、`git push origin v1.0.0` で push してください。

## 使い方

1. 起動すると設定ダイアログが開きます。
2. **「範囲を選択...」** をクリックし、画面に表示される半透明オーバーレイ上で **マウスドラッグ** してキャプチャしたい範囲を指定します（Esc でキャンセル）。
3. **保存先** に JPG/PDF を保存するフォルダを入力するか「参照...」で選択します。
4. **キー操作** に、1枚キャプチャするたびに送信するキーを指定します（例: `Enter`, `Tab`, `Ctrl+C`, `PageDown`）。
5. **最大枚数**（0 で無制限）と **「3枚連続同一で終了」** で終了条件を設定します。
6. **「開始」** を押すと、対象アプリをアクティブにした状態でキャプチャが始まります。
7. 終了後、指定フォルダに `screenshot_00001.jpg` … と `screenshots.pdf` が出力されます。

## 構成

- `main.go` — エントリ・設定ダイアログ起動・メインループ・PDF 出力
- `ui/dialog.go` — 設定ダイアログ（walk）
- `ui/region_select.go` — マウスで範囲選択するオーバーレイ（win32）
- `ui/folderbrowse_windows.go` — フォルダ選択ダイアログ（SHBrowseForFolder）
- `capture/capture.go` — 範囲キャプチャ（kbinani/screenshot）
- `keyboard/keyboard.go` — キー送信（sendinput）
- `compare/compare.go` — 画像ハッシュ・3枚同一判定
- `output/jpg.go` — JPG 保存
- `output/pdf.go` — JPG 一覧の PDF 化（gofpdf）
