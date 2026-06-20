package clipboard

import (
	"golang.design/x/clipboard"
)

func PasteClipboard() (string, error) {
	err := clipboard.Init()
	if err != nil {
		return "", err
	}

	return string(clipboard.Read(clipboard.FmtText)), nil
}

func CopyClipboard(text string) error {
	err := clipboard.Init()
	if err != nil {
		return err
	}

	_ = clipboard.Write(clipboard.FmtText, []byte(text))
	return nil
}
