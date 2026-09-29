// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

// filesServer records what sending files asks of Telegram.
type filesServer struct {
	uploads     int
	uploaded    []tg.InputMediaClass
	singles     []*tg.MessagesSendMediaRequest
	albums      []*tg.MessagesSendMultiMediaRequest
	nextPhotoID int64
}

func (f *filesServer) handle(in bin.Encoder) (bin.Encoder, error) {
	switch r := in.(type) {
	case *tg.UploadSaveFilePartRequest:
		f.uploads++
		return &tg.BoolTrue{}, nil
	case *tg.UploadSaveBigFilePartRequest:
		return &tg.BoolTrue{}, nil
	case *tg.MessagesUploadMediaRequest:
		f.uploaded = append(f.uploaded, r.Media)
		f.nextPhotoID++
		if _, ok := r.Media.(*tg.InputMediaUploadedPhoto); ok {
			return &tg.MessageMediaPhoto{Photo: &tg.Photo{ID: f.nextPhotoID, AccessHash: 7, FileReference: []byte{1}, Date: 1, DCID: 2}}, nil
		}
		return &tg.MessageMediaDocument{Document: &tg.Document{ID: 100 + f.nextPhotoID, AccessHash: 8, FileReference: []byte{2}, Date: 1, MimeType: "text/plain", DCID: 2}}, nil
	case *tg.MessagesSendMediaRequest:
		f.singles = append(f.singles, r)
		return &tg.Updates{}, nil
	case *tg.MessagesSendMultiMediaRequest:
		f.albums = append(f.albums, r)
		return &tg.Updates{}, nil
	}
	return nil, fmt.Errorf("unexpected %T", in)
}

func picture(t *testing.T, dir, name string, w, h int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 50, A: 255})
		}
	}
	var b bytes.Buffer
	if err := (&png.Encoder{CompressionLevel: png.NoCompression}).Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, b.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func filesStore(t *testing.T) (*Store, *filesServer) {
	s := testStore(t)
	s.history.peers[5] = peerRecord{Kind: "user", ID: 5, Hash: 55}
	server := new(filesServer)
	s.history.api = composerAPI(server.handle)
	return s, server
}

// Photos that can be grouped go in one album, in their order, the caption
// on the first of them; a picture too large for a photo is scaled down.
func TestSendFilesAsAnAlbum(t *testing.T) {
	s, server := filesStore(t)
	dir := t.TempDir()
	paths := []string{picture(t, dir, "a.png", 2000, 1000), picture(t, dir, "b.png", 30, 20), picture(t, dir, "c.png", 40, 40)}
	err := s.Send(context.Background(), 5, model.OutgoingMessage{RandomID: 1000, Text: " Holiday ", ReplyTo: 4,
		Files: &model.OutgoingFiles{Paths: paths, Group: true}})
	if err != nil {
		t.Fatal(err)
	}
	if len(server.singles) != 0 || len(server.albums) != 1 || server.uploads != 3 || len(server.uploaded) != 3 {
		t.Fatalf("%d singles, %d albums, %d uploads, %d uploaded", len(server.singles), len(server.albums), server.uploads, len(server.uploaded))
	}
	for i, media := range server.uploaded {
		if _, ok := media.(*tg.InputMediaUploadedPhoto); !ok {
			t.Fatalf("file %d uploaded as %T", i, media)
		}
	}
	album := server.albums[0]
	if reply, ok := album.ReplyTo.(*tg.InputReplyToMessage); !ok || reply.ReplyToMsgID != 4 {
		t.Fatalf("reply %+v", album.ReplyTo)
	}
	for i, item := range album.MultiMedia {
		photo, ok := item.Media.(*tg.InputMediaPhoto)
		if !ok || photo.ID.(*tg.InputPhoto).ID != int64(i+1) {
			t.Fatalf("item %d: %+v", i, item.Media)
		}
		if item.RandomID != 1000+int64(i) {
			t.Errorf("item %d has identity %d", i, item.RandomID)
		}
		if want := map[bool]string{true: "Holiday"}[i == 0]; item.Message != want {
			t.Errorf("item %d caption %q, want %q", i, item.Message, want)
		}
	}
	// Sending it again after a failure asks for the same identities.
	if err := s.Send(context.Background(), 5, model.OutgoingMessage{RandomID: 1000, Text: "Holiday", ReplyTo: 4, Files: &model.OutgoingFiles{Paths: paths, Group: true}}); err != nil {
		t.Fatal(err)
	}
	for i, item := range server.albums[1].MultiMedia {
		if item.RandomID != album.MultiMedia[i].RandomID {
			t.Fatalf("the second try has other identities: %d, %d", item.RandomID, album.MultiMedia[i].RandomID)
		}
	}
}

// Without grouping, each file is a message; the caption goes on the last,
// and a file that is no photo goes as a document.
func TestSendFilesOneByOne(t *testing.T) {
	s, server := filesStore(t)
	dir := t.TempDir()
	text := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(text, []byte("notes"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := s.Send(context.Background(), 5, model.OutgoingMessage{RandomID: 50, Text: "Look",
		Files: &model.OutgoingFiles{Paths: []string{picture(t, dir, "a.png", 30, 20), text}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(server.albums) != 0 || len(server.singles) != 2 {
		t.Fatalf("%d singles, %d albums", len(server.singles), len(server.albums))
	}
	if _, ok := server.singles[0].Media.(*tg.InputMediaUploadedPhoto); !ok || server.singles[0].Message != "" || server.singles[0].RandomID != 50 {
		t.Fatalf("the photo: %+v", server.singles[0])
	}
	doc, ok := server.singles[1].Media.(*tg.InputMediaUploadedDocument)
	if !ok || !doc.ForceFile || doc.MimeType != "text/plain" || server.singles[1].Message != "Look" || server.singles[1].RandomID != 51 {
		t.Fatalf("the document: %+v", server.singles[1])
	}
	if name := doc.Attributes[0].(*tg.DocumentAttributeFilename).FileName; name != "notes.txt" {
		t.Fatalf("named %q", name)
	}
}

// "Send as documents" keeps photos as they are, as files.
func TestSendPhotosAsDocuments(t *testing.T) {
	s, server := filesStore(t)
	path := picture(t, t.TempDir(), "raw.png", 3000, 2000)
	err := s.Send(context.Background(), 5, model.OutgoingMessage{RandomID: 9, Files: &model.OutgoingFiles{Paths: []string{path}, Documents: true}})
	if err != nil {
		t.Fatal(err)
	}
	doc, ok := server.singles[0].Media.(*tg.InputMediaUploadedDocument)
	if !ok || !doc.ForceFile || doc.MimeType != "image/png" {
		t.Fatalf("sent %+v", server.singles[0].Media)
	}
	if server.singles[0].Message != "" {
		t.Fatalf("a caption out of nowhere: %q", server.singles[0].Message)
	}
}

func TestSendFilesTellsWhichFileFailed(t *testing.T) {
	s, server := filesStore(t)
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty.png")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	err := s.Send(context.Background(), 5, model.OutgoingMessage{RandomID: 9, Files: &model.OutgoingFiles{Paths: []string{picture(t, dir, "ok.png", 8, 8), empty}}})
	if err == nil || !strings.Contains(err.Error(), "empty.png") {
		t.Fatalf("error %v", err)
	}
	if len(server.singles) != 0 || server.uploads != 0 {
		t.Fatal("something was sent before the files were checked")
	}
	if err := (model.OutgoingMessage{RandomID: 1, Files: &model.OutgoingFiles{}}).Validate(); err == nil {
		t.Fatal("no files validated")
	}
}
