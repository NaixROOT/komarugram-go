package diagnostics

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AutoExport owns a private directory per run. Only latest.json is replaced;
// shutdown also writes final.json. Disk use does not grow with each tick.
type AutoExport struct {
	recorder   *Recorder
	Directory  string
	interval   time.Duration
	stop, done chan struct{}
	once       sync.Once
}

func (r *Recorder) StartAutoExport(directory string, interval time.Duration) (*AutoExport, error) {
	if interval < 100*time.Millisecond {
		return nil, errors.New("profile-export interval must be at least 100ms")
	}
	r.exportMu.Lock()
	defer r.exportMu.Unlock()
	if r.exporter != nil {
		return nil, errors.New("automatic export already started")
	}
	if Current() != r {
		return nil, errors.New("profiler is closed")
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp(directory, "run-"+time.Now().Format("20060102-150405")+"-")
	if err != nil {
		return nil, err
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	e := &AutoExport{recorder: r, Directory: dir, interval: interval, stop: make(chan struct{}), done: make(chan struct{})}
	if err := e.write("latest.json"); err != nil {
		return nil, err
	}
	r.exporter = e
	go e.run()
	return e, nil
}
func (r *Recorder) AutoExporting() bool {
	r.exportMu.Lock()
	defer r.exportMu.Unlock()
	return r.exporter != nil
}
func (e *AutoExport) Stop() { e.once.Do(func() { close(e.stop) }); <-e.done }
func (e *AutoExport) run() {
	defer close(e.done)
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.report(e.write("latest.json"))
		case <-e.stop:
			e.report(e.write("latest.json"))
			e.report(e.write("final.json"))
			return
		}
	}
}
func (e *AutoExport) report(err error) {
	if err != nil {
		e.recorder.Event("profiler", "export.error", ErrorClass(err), 0, 0)
	}
}
func (e *AutoExport) write(name string) error {
	f, err := os.CreateTemp(e.Directory, ".snapshot-*")
	if err != nil {
		return err
	}
	path := f.Name()
	defer os.Remove(path)
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(e.recorder.Snapshot())
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(path, filepath.Join(e.Directory, name))
}
