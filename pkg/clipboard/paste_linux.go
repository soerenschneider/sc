package clipboard

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

var ErrToolNotFound = errors.New("tool not found")

func PasteClipboard() (string, error) {
	if _, err := exec.LookPath("wl-paste"); err == nil {
		cmd := exec.Command("wl-paste")
		out, err := cmd.Output()
		return string(out), err
	}

	return "", fmt.Errorf("%w: wl-paste", ErrToolNotFound)
}

func CopyClipboard(text string) error {
	// Prefer Wayland
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd := exec.Command("wl-copy")
			in, err := cmd.StdinPipe()
			if err != nil {
				return err
			}

			if err := cmd.Start(); err != nil {
				return err
			}

			_, err = in.Write([]byte(text))
			if err != nil {
				return err
			}
			_ = in.Close()

			return cmd.Wait()
		}
	}

	// Fallback to X11
	if _, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		in, err := cmd.StdinPipe()
		if err != nil {
			return err
		}

		if err := cmd.Start(); err != nil {
			return err
		}

		_, err = in.Write([]byte(text))
		if err != nil {
			return err
		}
		_ = in.Close()

		return cmd.Wait()
	}

	if _, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command("xsel", "--clipboard", "--input")
		in, err := cmd.StdinPipe()
		if err != nil {
			return err
		}

		if err := cmd.Start(); err != nil {
			return err
		}

		_, err = in.Write([]byte(text))
		if err != nil {
			return err
		}
		_ = in.Close()

		return cmd.Wait()
	}

	return errors.New("no clipboard utility found (install wl-copy or xclip)")
}
