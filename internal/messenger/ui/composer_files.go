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
	// form is the attachment chosen from the menu: 1 for a photo or a
	// video, 2 for a file. menu is set when the files are for the box
	// for sending files, which the attachment menu and the box's own
	// "Add" ask for.
	form int
	menu bool
	// path is the file chosen, the first when several were; paths are all.
	path  string
	paths []string
	err   error
}

// fileFilter narrows a file chooser to a kind of file: name is what the
// chooser calls it, extensions the files it shows, dot included.
type fileFilter struct {
	name       string
	extensions []string
}

// mediaFilter is what "Photo or video" shows: the images and the videos
// that can be sent as media.
var mediaFilter = fileFilter{"Media", []string{".jpg", ".jpeg", ".jfif", ".png", ".webp", ".gif", ".bmp", ".tif", ".tiff", ".mp4", ".m4v", ".webm", ".mov", ".mkv", ".avi", ".3gp"}}

func chooseAttachment(ctx context.Context, media bool) fileChoice {
	if media {
		return chooseFile(ctx, &mediaFilter)
	}
	return chooseFile(ctx, nil)
}

// chooseFile asks for a file in the system's own chooser, showing only the
// files of filter unless it is nil. A cancelled chooser chooses no path.
func chooseFile(ctx context.Context, filter *fileFilter) fileChoice {
	return chooseFiles(ctx, filter, false)
}

// chooseFiles is chooseFile that may take several files, as the attachment
// menu does. The chooser is the system's own: kdialog or zenity on Linux,
// the Windows and macOS dialogs elsewhere.
func chooseFiles(ctx context.Context, filter *fileFilter, several bool) fileChoice {
	var patterns []string
	if filter != nil {
		for _, ext := range filter.extensions {
			patterns = append(patterns, "*"+ext)
		}
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// The names come back as UTF-8, whatever the console's code page.
		script := `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; Add-Type -AssemblyName System.Windows.Forms; $dialog = New-Object System.Windows.Forms.OpenFileDialog; `
		if filter != nil {
			script += `$dialog.Filter = '` + filter.name + `|` + strings.Join(patterns, ";") + `'; `
		}
		if several {
			script += `$dialog.Multiselect = $true; `
		}
		script += `if ($dialog.ShowDialog() -eq 'OK') { $dialog.FileNames }`
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-STA", "-Command", script)
	case "darwin":
		var script string
		if several {
			script = `set chosen to choose file with multiple selections allowed`
		} else {
			script = `set chosen to {choose file`
		}
		if filter != nil {
			var types []string
			for _, ext := range filter.extensions {
				types = append(types, `"`+strings.TrimPrefix(ext, ".")+`"`)
			}
			script += ` of type {` + strings.Join(types, ", ") + `}`
		}
		if !several {
			script += `}`
		}
		script += `
set output to ""
repeat with one in chosen
set output to output & POSIX path of one & linefeed
end repeat
output`
		cmd = exec.CommandContext(ctx, "osascript", "-e", script)
	default:
		if _, err := exec.LookPath("kdialog"); err == nil {
			args := []string{"--getopenfilename", "."}
			if filter != nil {
				args = append(args, strings.Join(patterns, " ")+"|"+filter.name)
			}
			if several {
				args = append(args, "--multiple", "--separate-output")
			}
			cmd = exec.CommandContext(ctx, "kdialog", args...)
		} else if _, err := exec.LookPath("zenity"); err == nil {
			args := []string{"--file-selection"}
			if filter != nil {
				args = append(args, "--file-filter="+filter.name+" | "+strings.Join(patterns, " "))
			}
			if several {
				args = append(args, "--multiple", "--separator=\n")
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
	paths := splitPaths(string(out))
	choice := fileChoice{paths: paths}
	if len(paths) > 0 {
		choice.path = paths[0]
	}
	return choice
}

// splitPaths reads the paths a chooser printed, one to a line.
func splitPaths(out string) []string {
	var paths []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimRight(line, "\r"); strings.TrimSpace(line) != "" {
			paths = append(paths, line)
		}
	}
	return paths
}

// errNoChooser is a system without a file chooser the client can open.
var errNoChooser = errors.New("file chooser unavailable: enter the file path")
