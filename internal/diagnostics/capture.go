package diagnostics

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"runtime/trace"
	"time"
)

// Capture writes a private, opt-in local artifact. No HTTP/pprof listener is opened.
func (r *Recorder) Capture(directory, kind string) (string, error) {
	if !r.capture.CompareAndSwap(false, true) {
		return "", errors.New("another capture is running")
	}
	defer r.capture.Store(false)
	r.captureMu.Lock()
	defer r.captureMu.Unlock()
	if Current() != r {
		return "", errors.New("profiler is closed")
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", err
	}
	extension := ".pprof"
	if kind == "snapshot" {
		extension = ".json"
	}
	if kind == "trace" {
		extension = ".trace"
	}
	path, err := filepath.Abs(filepath.Join(directory, fmt.Sprintf("%s-%s%s", time.Now().Format("20060102-150405.000000000"), kind, extension)))
	if err != nil {
		return "", err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	defer file.Close()
	switch kind {
	case "snapshot":
		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		err = enc.Encode(r.Snapshot())
	case "cpu":
		if err = pprof.StartCPUProfile(file); err == nil {
			r.wait(15 * time.Second)
			pprof.StopCPUProfile()
		}
	case "trace":
		if err = trace.Start(file); err == nil {
			r.wait(5 * time.Second)
			trace.Stop()
		}
	case "heap", "allocs", "goroutine":
		err = pprof.Lookup(kind).WriteTo(file, 0)
	default:
		err = errors.New("unknown capture kind")
	}
	if err == nil {
		err = file.Sync()
	}
	return path, err
}
func (r *Recorder) wait(d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-r.done:
	}
}
