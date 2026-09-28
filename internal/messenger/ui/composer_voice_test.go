// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"gioui.org/io/key"

	"komarugram/internal/messenger/model"
)

// fakeRecorder has recorded a second of a loud tone.
type fakeRecorder struct {
	mu                 sync.Mutex
	stopped, cancelled bool
	failed             error
}

func (r *fakeRecorder) Level() float32          { return 0.5 }
func (r *fakeRecorder) Duration() time.Duration { return time.Second }
func (r *fakeRecorder) Full() bool              { return false }
func (r *fakeRecorder) Failed() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.failed
}
func (r *fakeRecorder) Stop() ([]int16, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stopped = true
	pcm := make([]int16, 48000)
	for i := range pcm {
		pcm[i] = int16(i%100*300 - 15000)
	}
	return pcm, nil
}
func (r *fakeRecorder) Cancel() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cancelled = true
}

func voiceHarness(t *testing.T) (*menuHarness, *[]*fakeRecorder) {
	h := newMenuHarness(t, nil)
	var recorders []*fakeRecorder
	h.page.composer.voice = voiceTools{
		record: func(context.Context, string) (voiceRecorder, error) {
			r := &fakeRecorder{}
			recorders = append(recorders, r)
			return r, nil
		},
		encode: func(_ context.Context, _ string, _ []int16, path string) error {
			return os.WriteFile(path, []byte("OggS"), 0o600)
		},
	}
	return h, &recorders
}

// The microphone records while the text is empty; Send sends the
// recording as a voice message, whose file goes once it is sent.
func TestComposerRecordsVoice(t *testing.T) {
	h, recorders := voiceHarness(t)
	c := h.page.composer
	if !c.canRecord(c.draft(1)) {
		t.Fatal("no microphone with an empty composer")
	}
	c.micClick.click.Click()
	h.frames(3)
	if c.recording == nil || len(*recorders) != 1 {
		t.Fatal("the microphone did not record")
	}
	c.send.click.Click()
	h.frames(2)
	var sent []model.OutgoingMessage
	for deadline := time.Now().Add(5 * time.Second); len(sent) == 0; {
		if time.Now().After(deadline) {
			t.Fatal("nothing sent")
		}
		h.frame()
		h.store.mu.Lock()
		sent = append([]model.OutgoingMessage(nil), h.store.sent...)
		h.store.mu.Unlock()
	}
	m := sent[0]
	if m.Voice == nil || m.Voice.Duration != time.Second || len(m.Voice.Waveform) != 63 || m.Path == "" || !(*recorders)[0].stopped {
		t.Fatalf("sent %+v", m)
	}
	for deadline := time.Now().Add(5 * time.Second); ; {
		h.frame()
		if _, err := os.Stat(m.Path); errors.Is(err, os.ErrNotExist) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the recording's file was kept after it was sent")
		}
	}
	// Text in the composer takes the microphone's place.
	c.draft(1).editor.SetText("hi")
	if c.canRecord(c.draft(1)) {
		t.Fatal("microphone offered beside text")
	}
}

// Escape drops the recording; a microphone that fails tells so.
func TestComposerCancelsVoice(t *testing.T) {
	h, recorders := voiceHarness(t)
	c := h.page.composer
	// The history has the keyboard, as after a click in it; it takes
	// Escape to clear its selection.
	h.router.Source().Execute(key.FocusCmd{Tag: &h.page.keyboard})
	h.frame()
	c.micClick.click.Click()
	h.frames(3)
	h.router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
	h.frames(3)
	if c.recording != nil {
		t.Fatal("Escape left the recording")
	}
	if !h.router.Source().Focused(&c.draft(1).editor) {
		t.Fatal("the field did not get the focus back")
	}
	for deadline := time.Now().Add(5 * time.Second); ; {
		(*recorders)[0].mu.Lock()
		cancelled := (*recorders)[0].cancelled
		(*recorders)[0].mu.Unlock()
		if cancelled {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the recorder was not cancelled")
		}
		time.Sleep(5 * time.Millisecond)
	}
	c.micClick.click.Click()
	h.frames(3)
	r := (*recorders)[1]
	r.mu.Lock()
	r.failed = errors.New("no microphone")
	r.mu.Unlock()
	h.frames(2)
	if c.recording != nil || c.draft(1).err == nil {
		t.Fatal("a failed recording went on")
	}
	h.store.mu.Lock()
	defer h.store.mu.Unlock()
	if len(h.store.sent) != 0 {
		t.Fatalf("sent %+v", h.store.sent)
	}
}
