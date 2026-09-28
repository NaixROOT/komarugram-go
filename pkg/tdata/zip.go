// SPDX-License-Identifier: Unlicense

package tdata

import (
	"archive/zip"
	"bytes"
	"io"
	"komarugram/pkg/helpers"
	"komarugram/pkg/set"
	"os"
	"path/filepath"
)

func ReadZip(r io.ReaderAt, size int64) ([]TDataSession, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}

	parents := map[string]*TDataParent{}
	dirsMap := map[string]*set.Set[string]{}

	for _, entry := range zr.File {
		dir, name := filepath.Split(entry.Name)
		parentPath, parentDir := filepath.Split(helpers.ExcludeLastChar(dir))

		dirSet, ok := dirsMap[parentPath]
		if len(parentDir) == sessionFileNameSize {
			if !ok {
				dirSet = set.NewSet[string]()
				dirsMap[parentPath] = dirSet
			}
			dirSet.Add(parentDir)
		}

		if isTdesktopKey(name) {
			parent, ok := parents[dir]
			if ok {
				continue
			}
			parent = new(TDataParent)
			parents[dir] = parent

			keyFile, err := entry.Open()
			if err != nil {
				return nil, err
			}
			keyData, err := io.ReadAll(keyFile)
			keyFile.Close()
			if err != nil {
				return nil, err
			}
			parent.tdesktopKey, err = readTDesktopKey(keyData)
			if err != nil {
				return nil, err
			}
		}
	}

	var sessionsCount int
	for _, dirSet := range dirsMap {
		sessionsCount += dirSet.Size()
	}

	results := make([]TDataSession, 0, sessionsCount)

	for _, entry := range zr.File {
		dir, name := filepath.Split(entry.Name)

		dirSet, ok := dirsMap[dir]
		if !ok {
			continue
		}

		parent, ok := parents[dir]
		if !ok {
			continue
		}

		if len(name) == sessionFileNameSize+1 && dirSet.Has(name[:sessionFileNameSize]) {
			sessionFile, err := entry.Open()
			if err != nil {
				return results, err
			}
			sessionData, err := io.ReadAll(sessionFile)
			sessionFile.Close()
			if err != nil {
				return results, err
			}
			result, err := readTDesktopSession(sessionData, parent.tdesktopKey)
			if err != nil {
				continue
			}
			results = append(results, result)
		}
	}

	return results, err
}

func ReadZipBytes(data []byte) ([]TDataSession, error) {
	return ReadZip(bytes.NewReader(data), int64(len(data)))
}

func WriteZip(zipPath string, sessions []TDataSession) error {
	files, err := buildTDesktopSessionFiles(sessions)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(zipPath), tdesktopDirMode); err != nil {
		return err
	}

	file, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return writeTDesktopSessionZip(file, files)
}

func WriteZipTo(w io.Writer, sessions []TDataSession) error {
	files, err := buildTDesktopSessionFiles(sessions)
	if err != nil {
		return err
	}
	return writeTDesktopSessionZip(w, files)
}

func CreateZipBytes(sessions []TDataSession) ([]byte, error) {
	var b bytes.Buffer
	err := WriteZipTo(&b, sessions)
	return b.Bytes(), err
}
