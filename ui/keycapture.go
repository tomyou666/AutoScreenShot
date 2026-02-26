//go:build windows

package ui

import (
	"strings"

	"github.com/lxn/walk"
)

// keyDisplayName は keyboard.Send で使う名前（表示・送信用）に変換します。
// walk の Key.String() は "Return" を返すが、設定例は "Enter" のため統一する。
var keyDisplayNames = map[string]string{
	"Return": "Enter",
	"Right":  "ArrowRight",
	"Left":   "ArrowLeft",
	"Up":     "ArrowUp",
	"Down":   "ArrowDown",
}

// keyOperationString は、現在押されている修飾キーと押されたキーから
// 「キー操作」欄に設定する文字列（例: "Enter", "Ctrl+C"）を組み立てます。
func keyOperationString(mod walk.Modifiers, key walk.Key) string {
	// 修飾キーだけの押下は無視（メインキーが必要）
	if isModifierOnly(key) {
		return ""
	}
	keyStr := key.String()
	if d, ok := keyDisplayNames[keyStr]; ok {
		keyStr = d
	} else {
		keyStr = strings.ToUpper(keyStr)
	}
	var parts []string
	if mod&walk.ModControl != 0 {
		parts = append(parts, "Ctrl")
	}
	if mod&walk.ModAlt != 0 {
		parts = append(parts, "Alt")
	}
	if mod&walk.ModShift != 0 {
		parts = append(parts, "Shift")
	}
	parts = append(parts, keyStr)
	return strings.Join(parts, "+")
}

func isModifierOnly(key walk.Key) bool {
	switch key {
	case walk.KeyShift, walk.KeyControl, walk.KeyAlt,
		walk.KeyLShift, walk.KeyRShift, walk.KeyLControl, walk.KeyRControl,
		walk.KeyLAlt, walk.KeyRAlt,
		walk.KeyLWin, walk.KeyRWin:
		return true
	}
	return false
}
