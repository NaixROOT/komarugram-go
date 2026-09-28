// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

type fileChoice struct {
	chat int64
	form int
	path string
	err  error
}

func chooseAttachment(ctx context.Context, media bool) fileChoice {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-STA", "-Command", `Add-Type -AssemblyName System.Windows.Forms; $dialog = New-Object System.Windows.Forms.OpenFileDialog; if ($dialog.ShowDialog() -eq 'OK') { $dialog.FileName }`)
	case "darwin":
		cmd = exec.CommandContext(ctx, "osascript", "-e", `POSIX path of (choose file)`)
	default:
		if _, err := exec.LookPath("kdialog"); err == nil {
			args := []string{"--getopenfilename", "."}
			if media {
				args = append(args, "*.jpg *.jpeg *.png *.webp *.gif *.mp4 *.webm *.mov")
			}
			cmd = exec.CommandContext(ctx, "kdialog", args...)
		} else if _, err := exec.LookPath("zenity"); err == nil {
			args := []string{"--file-selection"}
			if media {
				args = append(args, "--file-filter=Media | *.jpg *.jpeg *.png *.webp *.gif *.mp4 *.webm *.mov")
			}
			cmd = exec.CommandContext(ctx, "zenity", args...)
		} else {
			return fileChoice{err: errors.New("file chooser unavailable: enter the file path")}
		}
	}
	out, err := cmd.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return fileChoice{}
		}
		return fileChoice{err: err}
	}
	return fileChoice{path: strings.TrimSpace(string(out))}
}
