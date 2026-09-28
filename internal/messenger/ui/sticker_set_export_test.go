// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"komarugram/internal/messenger/model"
)

type stickerArchiveTestSource map[string][]byte

func (s stickerArchiveTestSource) Media(ctx context.Context, msg model.Message) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s[msg.Media.ID], nil
}

func TestStickerSetArchiveContainsOriginalDocuments(t *testing.T) {
	pack := model.StickerSet{Title: "Rigby chan", Items: []model.PickerItem{
		{Emoji: "🐈", DocumentID: 11, Media: model.Message{Kind: model.MessageSticker, Media: &model.MessageMedia{ID: "webm", MIMEType: "video/webm"}}},
		{Emoji: "🍒", DocumentID: 12, Media: model.Message{Kind: model.MessageSticker, Media: &model.MessageMedia{ID: "tgs", MIMEType: "application/x-tgsticker"}}},
	}}
	source := stickerArchiveTestSource{"webm": []byte("original webm bytes"), "tgs": []byte("original tgs bytes")}
	path := filepath.Join(t.TempDir(), "pack")
	if err := writeStickerSetArchive(context.Background(), source, pack, path); err != nil {
		t.Fatal(err)
	}
	archive, err := zip.OpenReader(path + ".zip")
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	want := map[string][]byte{"001.webm": source["webm"], "002.tgs": source["tgs"]}
	var manifest struct {
		Title string               `json:"title"`
		Items []stickerArchiveItem `json:"items"`
	}
	for _, file := range archive.File {
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		contents, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		if file.Name == "manifest.json" {
			if err := json.Unmarshal(contents, &manifest); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if !bytes.Equal(contents, want[file.Name]) || want[file.Name] == nil {
			t.Fatalf("%s does not contain its original bytes", file.Name)
		}
		delete(want, file.Name)
	}
	if len(want) != 0 || manifest.Title != pack.Title || len(manifest.Items) != 2 || manifest.Items[0].File != "001.webm" || manifest.Items[1].DocumentID != 12 {
		t.Fatalf("archive contents or manifest are incomplete: remaining=%v manifest=%+v", want, manifest)
	}
}

func TestStickerSetArchiveCancellationLeavesNoFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	path := filepath.Join(t.TempDir(), "cancelled.zip")
	pack := model.StickerSet{Items: []model.PickerItem{{Media: model.Message{Media: &model.MessageMedia{ID: "webm", MIMEType: "video/webm"}}}}}
	if err := writeStickerSetArchive(ctx, stickerArchiveTestSource{"webm": []byte("data")}, pack, path); err == nil {
		t.Fatal("cancelled export succeeded")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("cancelled archive remains: %v", err)
	}
}

func TestStickerSetArchiveFailurePreservesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pack.zip")
	if err := os.WriteFile(path, []byte("previous archive"), 0o600); err != nil {
		t.Fatal(err)
	}
	pack := model.StickerSet{Items: []model.PickerItem{
		{Media: model.Message{Media: &model.MessageMedia{ID: "first", MIMEType: "video/webm"}}},
		{Media: model.Message{Media: &model.MessageMedia{ID: "missing", MIMEType: "video/webm"}}},
	}}
	if err := writeStickerSetArchive(context.Background(), stickerArchiveTestSource{"first": []byte("data")}, pack, path); err == nil {
		t.Fatal("incomplete export succeeded")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "previous archive" {
		t.Fatalf("existing archive was changed: %q, %v", got, err)
	}
}

func TestStickerArchiveNameIsSafe(t *testing.T) {
	if got := stickerArchiveName("Rigby / Chan"); got != "Rigby_Chan.zip" {
		t.Fatal(got)
	}
}
