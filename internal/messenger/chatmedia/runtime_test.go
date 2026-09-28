// SPDX-License-Identifier: Unlicense

package chatmedia

import (
	"context"
	"testing"
)

type fakeRuntime struct{ closed int }

func (r *fakeRuntime) Close(context.Context) error { r.closed++; return nil }

func TestCodecRetireWaitsForUsers(t *testing.T) {
	var opened []*fakeRuntime
	c := codec[*fakeRuntime]{open: func(context.Context) (*fakeRuntime, error) {
		r := new(fakeRuntime)
		opened = append(opened, r)
		return r, nil
	}}
	first, release, err := c.acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if again, releaseAgain, _ := c.acquire(context.Background()); again != first {
		t.Fatal("a second worker compiled another runtime")
	} else {
		releaseAgain()
	}
	c.retire()
	if first.closed != 0 {
		t.Fatal("retire closed a runtime still in use")
	}
	next, releaseNext, _ := c.acquire(context.Background())
	if next == first {
		t.Fatal("a retired runtime was handed out again")
	}
	release()
	release()
	if first.closed != 1 {
		t.Fatalf("retired runtime closed %d times after its last user, want 1", first.closed)
	}
	releaseNext()
	if next.closed != 0 {
		t.Fatal("releasing the current runtime closed it")
	}
	c.retire()
	if next.closed != 1 || len(opened) != 2 {
		t.Fatalf("idle runtime closed %d times, %d opened", next.closed, len(opened))
	}
}
