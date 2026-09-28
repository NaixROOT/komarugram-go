package h264

import (
	"bytes"
	"context"
	"image"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestZZVideo(t *testing.T) {
	ctx := context.Background()
	data, _ := os.ReadFile("../../assets/video.mp4")
	start := time.Now()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close(ctx)
	t.Logf("compile %v", time.Since(start))
	v, err := rt.OpenVideo(ctx, data)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close(ctx)
	// ffmpeg's first 10 pictures, planar.
	out, err := exec.Command("ffmpeg", "-loglevel", "error", "-i", "../../assets/video.mp4", "-frames:v", "10", "-f", "rawvideo", "-pix_fmt", "yuv420p", "-").Output()
	if err != nil {
		t.Fatal(err)
	}
	w, h := 720, 1280
	frame := w * h * 3 / 2
	for i := range 10 {
		idx, err := v.Next(ctx)
		if err != nil {
			t.Fatal(err)
		}
		p := v.picture
		var got bytes.Buffer
		for y := range h {
			got.Write(p.Y[y*p.YStride : y*p.YStride+w])
		}
		for y := range h / 2 {
			got.Write(p.Cb[y*p.CStride : y*p.CStride+w/2])
		}
		for y := range h / 2 {
			got.Write(p.Cr[y*p.CStride : y*p.CStride+w/2])
		}
		if !bytes.Equal(got.Bytes(), out[i*frame:(i+1)*frame]) {
			t.Fatalf("picture %d (index %d) differs from ffmpeg", i, idx)
		}
	}
	t.Log("10 pictures identical to ffmpeg")
	start = time.Now()
	n := 0
	var pic time.Duration
	for n < 1081+5 {
		if _, err := v.Next(ctx); err != nil {
			t.Fatal(err)
		}
		s := time.Now()
		v.Picture(image.Pt(216, 384))
		pic += time.Since(s)
		n++
	}
	took := time.Since(start)
	t.Logf("%d pictures through a loop: %.2f ms each, of it scaling to 216x384 %.2f ms", n, float64(took.Microseconds())/float64(n)/1000, float64(pic.Microseconds())/float64(n)/1000)
	for _, size := range []image.Point{{100, 178}, {400, 711}, {720, 1280}} {
		s := time.Now()
		for range 30 {
			v.Picture(size)
		}
		t.Logf("scaling to %v: %.2f ms", size, float64(time.Since(s).Microseconds())/30/1000)
	}
}
