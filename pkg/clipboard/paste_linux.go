package clipboard

import (
	"os/exec"
)

func PasteClipboard() (string, error) {
	cmd := exec.Command("wl-paste")
	out, err := cmd.Output()

	return string(out), err
}
