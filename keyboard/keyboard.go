//go:build windows

package keyboard

import (
	"strings"

	"github.com/dacapoday/sendinput"
)

// Send はキー操作文字列（例: "Enter", "Tab", "Ctrl+C"）を1回送信します。
func Send(keyOperation string) error {
	keyOperation = strings.TrimSpace(keyOperation)
	if keyOperation == "" {
		return nil
	}
	parts := strings.Split(keyOperation, "+")
	var modifiers []sendinput.KeyCode
	for i, p := range parts {
		p = strings.TrimSpace(p)
		parts[i] = strings.ToUpper(p)
	}
	var mainKey string
	for i, p := range parts {
		if i < len(parts)-1 {
			switch p {
			case "CTRL", "CONTROL":
				modifiers = append(modifiers, sendinput.KEY_LCONTROL)
			case "ALT":
				modifiers = append(modifiers, sendinput.KEY_LMENU)
			case "SHIFT":
				modifiers = append(modifiers, sendinput.KEY_LSHIFT)
			case "WIN":
				modifiers = append(modifiers, sendinput.KEY_LWIN)
			}
		} else {
			mainKey = p
		}
	}
	if mainKey == "" {
		mainKey = strings.TrimSpace(strings.ToUpper(keyOperation))
	}
	main := sendinput.Key(mainKey)
	if main == 0 && len(mainKey) == 1 {
		main = sendinput.KeyCode(mainKey[0])
		if main >= 0x41 && main <= 0x5A {
			// already correct
		} else if main >= 0x30 && main <= 0x39 {
			// digit
		} else {
			main = sendinput.Key(mainKey)
		}
	}
	// 修飾キーを押す
	for _, m := range modifiers {
		_ = sendinput.SendKeyboardInput(m, true)
	}
	// メインキーを押して離す
	if err := sendinput.SendKeyboardInput(main, true); err != nil {
		releaseModifiers(modifiers)
		return err
	}
	if err := sendinput.SendKeyboardInput(main, false); err != nil {
		releaseModifiers(modifiers)
		return err
	}
	// 修飾キーを離す（逆順）
	releaseModifiers(modifiers)
	return nil
}

func releaseModifiers(modifiers []sendinput.KeyCode) {
	for i := len(modifiers) - 1; i >= 0; i-- {
		_ = sendinput.SendKeyboardInput(modifiers[i], false)
	}
}
