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

// fileFilter narrows a file chooser to a kind of file: name is what the
// chooser calls it, extensions the files it shows, dot included.
type fileFilter struct {
	name       string
	extensions []string
}

var mediaFilter = fileFilter{"Media", []string{".jpg", ".jpeg", ".png", ".webp", ".gif", ".mp4", ".webm", ".mov"}}

func chooseAttachment(ctx context.Context, media bool) fileChoice {
	if media {
		return chooseFile(ctx, &mediaFilter)
	}
	return chooseFile(ctx, nil)
}

// chooseFile asks for a file in the system's own chooser, showing only the
// files of filter unless it is nil. A cancelled chooser chooses no path.
func chooseFile(ctx context.Context, filter *fileFilter) fileChoice {
	var patterns []string
	if filter != nil {
		for _, ext := range filter.extensions {
			patterns = append(patterns, "*"+ext)
		}
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		script := `Add-Type -AssemblyName System.Windows.Forms; $dialog = New-Object System.Windows.Forms.OpenFileDialog; `
		if filter != nil {
			script += `$dialog.Filter = '` + filter.name + `|` + strings.Join(patterns, ";") + `'; `
		}
		script += `if ($dialog.ShowDialog() -eq 'OK') { $dialog.FileName }`
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-STA", "-Command", script)
	case "darwin":
		script := `POSIX path of (choose file`
		if filter != nil {
			var types []string
			for _, ext := range filter.extensions {
				types = append(types, `"`+strings.TrimPrefix(ext, ".")+`"`)
			}
			script += ` of type {` + strings.Join(types, ", ") + `}`
		}
		cmd = exec.CommandContext(ctx, "osascript", "-e", script+`)`)
	default:
		if _, err := exec.LookPath("kdialog"); err == nil {
			args := []string{"--getopenfilename", "."}
			if filter != nil {
				args = append(args, strings.Join(patterns, " ")+"|"+filter.name)
			}
			cmd = exec.CommandContext(ctx, "kdialog", args...)
		} else if _, err := exec.LookPath("zenity"); err == nil {
			args := []string{"--file-selection"}
			if filter != nil {
				args = append(args, "--file-filter="+filter.name+" | "+strings.Join(patterns, " "))
			}
			cmd = exec.CommandContext(ctx, "zenity", args...)
		} else {
			return fileChoice{err: errNoChooser}
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

// errNoChooser is a system without a file chooser the client can open.
var errNoChooser = errors.New("file chooser unavailable: enter the file path")
