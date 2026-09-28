// Package crash turns panics in isolated units of work — one UI frame, one
// media load — into logged reports instead of a process exit.
//
// Only code whose state is discarded or rebuilt after the failure should be
// guarded: a recovered panic leaves whatever it was mutating half done.
package crash

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"
)

// Panic is a recovered panic with the stack of the goroutine that raised it.
type Panic struct {
	Where string
	Value any
	Stack []byte
	At    time.Time
	Path  string
}

func (p *Panic) Error() string { return fmt.Sprintf("panic in %s: %v", p.Where, p.Value) }

// Text is the complete report offered by the copy button, even if writing a
// file failed.
func (p *Panic) Text() string {
	at := p.At
	if at.IsZero() {
		at = time.Now()
	}
	return fmt.Sprintf("time: %s\nwhere: %s\npanic: %v\n\n%s", at.Format(time.RFC3339Nano), p.Where, p.Value, p.Stack)
}

// Guard runs fn and returns the recovered panic, if any. The panic is
// reported before Guard returns.
func Guard(where string, fn func()) (p *Panic) {
	defer func() {
		if v := recover(); v != nil {
			p = &Panic{Where: where, Value: v, Stack: debug.Stack()}
			Report(p)
		}
	}()
	fn()
	return nil
}

// Recover is deferred at the top of a goroutine. It reports a panic and,
// when handle is not nil, passes it on so the owner can mark its work failed.
func Recover(where string, handle func(*Panic)) {
	v := recover()
	if v == nil {
		return
	}
	p := &Panic{Where: where, Value: v, Stack: debug.Stack()}
	Report(p)
	if handle != nil {
		handle(p)
	}
}

const (
	maxReports = 32 // files kept in the report directory
	reportGap  = 10 * time.Second
)

var (
	mu        sync.Mutex
	lastWrite = map[string]time.Time{}
	lastPath  = map[string]string{}
	subs      = map[uint64]func(*Panic){}
	nextID    uint64
)

// Subscribe receives reported panics after the report path has been set.
// A notifier must return promptly; it may open UI asynchronously.
func Subscribe(notify func(*Panic)) func() {
	mu.Lock()
	id := nextID
	nextID++
	subs[id] = notify
	mu.Unlock()
	return func() { mu.Lock(); delete(subs, id); mu.Unlock() }
}

func notify(p *Panic) {
	mu.Lock()
	callbacks := make([]func(*Panic), 0, len(subs))
	for _, callback := range subs {
		callbacks = append(callbacks, callback)
	}
	mu.Unlock()
	for _, callback := range callbacks {
		func() {
			defer func() {
				if v := recover(); v != nil {
					log.Printf("crash notifier failed: %v", v)
				}
			}()
			callback(p)
		}()
	}
}

// Report logs a panic and writes it to <user cache>/komarugram-go/crashes. A panic
// that repeats every frame is written at most once per reportGap per site so
// it cannot fill the disk; the log still gets a one-line note each time.
func Report(p *Panic) {
	p.At = time.Now()
	defer notify(p)
	mu.Lock()
	due := time.Since(lastWrite[p.Where]) >= reportGap
	p.Path = lastPath[p.Where]
	if due {
		lastWrite[p.Where] = time.Now()
	}
	mu.Unlock()
	if !due {
		log.Printf("recovered %v (repeat, report suppressed)", p)
		return
	}
	path, err := write(p)
	if err != nil {
		log.Printf("recovered %v\n%s(crash report not saved: %v)", p, p.Stack, err)
		return
	}
	p.Path = path
	mu.Lock()
	lastPath[p.Where] = path
	mu.Unlock()
	log.Printf("recovered %v\n%sreport: %s", p, p.Stack, path)
}

func Dir() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cache, "komarugram-go", "crashes"), nil
}

func write(p *Panic) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	now := p.At
	name := fmt.Sprintf("panic-%s-%09d.txt", now.Format("20060102-150405"), now.Nanosecond())
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(p.Text()), 0o600); err != nil {
		return "", err
	}
	prune(dir)
	return path, nil
}

// prune keeps the newest maxReports files; names sort by time.
func prune(dir string) {
	matches, _ := filepath.Glob(filepath.Join(dir, "panic-*.txt"))
	for len(matches) > maxReports {
		os.Remove(matches[0])
		matches = matches[1:]
	}
}
