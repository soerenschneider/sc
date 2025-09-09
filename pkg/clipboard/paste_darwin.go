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
