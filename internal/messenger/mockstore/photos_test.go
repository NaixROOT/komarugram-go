package mockstore

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"testing"
	"time"
)

func TestDemoPhotosAndGallery(t *testing.T) {
	s := New(time.Now(), 0)
	page, e := s.ChatPhotos(context.Background(), 2, 1<<30, -1, 100)
	photos := page.Messages
	if e != nil || len(photos) < 10 || page.Total != len(photos) || page.More {
		t.Fatalf("%d photos, %v", len(photos), e)
	}
	first := photos[0].Media
	if len(first.Variants) != 3 || first.Variants[0].Width != 320 {
		t.Fatalf("variants %+v", first.Variants)
	}
	for _, m := range append(first.Variants[:1:1], *first) {
		data, e := s.Media(context.Background(), photos[0].WithMedia(&m))
		if e != nil {
			t.Fatal(e)
		}
		cfg, e := jpeg.DecodeConfig(bytes.NewReader(data))
		if e != nil || image.Pt(cfg.Width, cfg.Height) != image.Pt(m.Width, m.Height) {
			t.Fatalf("%s decodes as %dx%d: %v", m.ID, cfg.Width, cfg.Height, e)
		}
	}
	newer, _ := s.ChatPhotos(context.Background(), 2, photos[2].Key.MessageID, 1, 2)
	if len(newer.Messages) != 2 || newer.Messages[0].Key != photos[3].Key || !newer.More {
		t.Fatal("newer page")
	}
}
