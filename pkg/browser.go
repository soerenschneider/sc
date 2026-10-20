package pkg

import (
	"os/exec"
	"runtime"
)

// OpenBrowser opens url in the default browser without waiting for it.
func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default: // linux, etc.
		return exec.Command("xdg-open", url).Start()
	}
}
