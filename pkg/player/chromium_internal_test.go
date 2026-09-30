// SPDX-License-Identifier: Unlicense OR MIT

package player

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"komarugram/pkg/miniapp"
)

// TestChromiumFit checks that the window takes the shape of the video once
// the video tells its size.
func TestChromiumFit(t *testing.T) {
	if !miniapp.Available() {
		t.Skip("no Chromium-based browser")
	}
	path := os.Getenv("PLAYER_VIDEO")
	if path == "" {
		paths, _ := filepath.Glob("../../assets/rigby/*.webm")
		if len(paths) == 0 {
			t.Skip("no PLAYER_VIDEO or ../../assets/rigby/*.webm to play")
		}
		path = paths[0]
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	stream, err := Serve("video", file, info.Size())
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	p, err := openChromium(context.Background(), stream.URL(), []string{"--headless=new", "--mute-audio"})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	var size struct{ VW, VH, W, H float64 }
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		answer, err := p.page.Eval(context.Background(), `(function () {
		  var v = document.getElementById('player');
		  return JSON.stringify({vw: v.videoWidth, vh: v.videoHeight, w: innerWidth, h: innerHeight});
		})()`)
		if err == nil && json.Unmarshal([]byte(answer), &size) == nil && size.VW > 0 && size.W > 0 &&
			math.Abs(size.W/size.H-size.VW/size.VH) < 0.02 {
			t.Logf("video %vx%v, window %vx%v", size.VW, size.VH, size.W, size.H)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("the window never took the shape of the video: %+v", size)
}
