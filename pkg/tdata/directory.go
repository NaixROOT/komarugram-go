// SPDX-License-Identifier: Unlicense

package tdata

import (
	"os"
	"path/filepath"
	"slices"
)

func ReadDir(tdataPath string) ([]TDataSession, error) {
	entries, err := os.ReadDir(tdataPath)
	if err != nil {
		return nil, err
	}

	var tdataKey []byte
	var sessions []string
	var dirs []string

	for _, entry := range entries {
		name := entry.Name()
		if len(tdataKey) == 0 && isTdesktopKey(name) {
			keyFile, err := os.ReadFile(filepath.Join(tdataPath, name))
			if err != nil {
				return nil, err
			}
			tdataKey, err = readTDesktopKey(keyFile)
		} else if entry.IsDir() {
			if len(name) == sessionFileNameSize {
				sessions = append(sessions, name)
			} else {
				dirs = append(dirs, name)
			}
		}
	}

	var results []TDataSession

	if len(tdataKey) > 0 {
		results = make([]TDataSession, 0, len(sessions))
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if len(name) == sessionFileNameSize+1 && slices.Contains(sessions, name[:sessionFileNameSize]) {
				sessionFile, err := os.ReadFile(filepath.Join(tdataPath, name))
				if err != nil {
					return results, err
				}
				result, err := readTDesktopSession(sessionFile, tdataKey)
				if err != nil {
					continue
				}
				results = append(results, result)
			}

		}
	} else {
		results = []TDataSession{}
		for _, dir := range dirs {
			nextResults, err := ReadDir(filepath.Join(tdataPath, dir))
			if err != nil {
				return results, err
			}
			results = append(results, nextResults...)
		}
	}

	return results, err
}

func WriteDir(tdataPath string, sessions []TDataSession) error {
	files, err := buildTDesktopSessionFiles(sessions)
	if err != nil {
		return err
	}
	return writeTDesktopSessionFiles(tdataPath, files)
}
