// SPDX-License-Identifier: Unlicense OR MIT

package miniapp

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"komarugram/pkg/program"
)

// BrowserEnv names the browser to drive, overriding the one that would
// otherwise be chosen: a program on PATH, an absolute path to one, or the
// application id of an installed flatpak.
const BrowserEnv = "KITCHEN_MINIAPP_BROWSER"

// nativeBrowsers are the programs looked for on PATH. Among browsers of the
// same version, the one listed first wins.
var nativeBrowsers = []string{
	"chromium", "chromium-browser",
	"google-chrome", "google-chrome-stable",
	"brave-browser", "brave",
}

// flatpakBrowsers are the Chromium-based flatpaks looked for. They rank after
// the native browsers when versions tie.
var flatpakBrowsers = []string{
	"io.github.ungoogled_software.ungoogled_chromium",
	"com.github.Eloston.UngoogledChromium",
	"org.chromium.Chromium",
	"com.google.Chrome",
	"com.brave.Browser",
}

// browser is the browser this client drives: either a program on this machine
// or a flatpak to run one from.
type browser struct {
	ref     string // a path, or an application id
	flatpak bool
	found   bool
	// version is the Chromium the browser is built on, as far as it tells;
	// nil when it could not be read.
	version version
	// banner is what the browser printed for --version.
	banner string
}

// command returns what to run to open a browser on profile.
func (b browser) command(profile string) (string, []string) {
	if !b.flatpak {
		return b.ref, nil
	}
	// A flatpak sees only what it is given: /tmp it already has, but a profile
	// anywhere else — a cache directory, say — is invisible to it, and the
	// browser would quietly keep its own copy inside the sandbox, where nothing
	// here could read it. The directory is granted by name, for this run only.
	pre := []string{"run"}
	if profile != "" {
		pre = append(pre, "--filesystem="+profile)
	}
	return "flatpak", append(pre, b.ref)
}

// Available reports whether a browser to drive was found.
func Available() bool { return findBrowser().found }

// Browser names what a launch will run: the path of a program, or the
// application id of a flatpak. It is empty when nothing was found.
func Browser() string {
	found := findBrowser()
	if !found.found {
		return ""
	}
	return found.ref
}

// BrowserVersion is what the browser a launch will run printed for
// --version, such as "Chromium 152.0.7977.82". It is empty when nothing was
// found or the browser would not say.
func BrowserVersion() string { return findBrowser().banner }

var (
	browserOnce  sync.Once
	browserFound browser

	// custom is the browser the user picked, and customFound what it said
	// for --version, asked once per path.
	customMu     sync.Mutex
	custom       string
	customProbed string
	customFound  browser
)

// SetBrowser makes launches run the browser at path, which CheckBrowser
// accepted, instead of the one found; "" goes back to finding one.
// BrowserEnv still wins over it. A path that no longer answers as a
// Chromium-based browser is passed over for the one found.
func SetBrowser(path string) {
	customMu.Lock()
	custom = path
	customMu.Unlock()
}

// customBrowser returns the browser the user picked, found only while it
// still answers as a Chromium-based one.
func customBrowser() browser {
	customMu.Lock()
	defer customMu.Unlock()
	if custom == "" {
		return browser{}
	}
	if customProbed != custom {
		customProbed = custom
		customFound = browser{}
		if banner, err := CheckBrowser(context.Background(), custom); err == nil {
			customFound = browserAt(custom)
			customFound.banner = banner
			customFound.version = parseVersion(banner)
		}
	}
	return customFound
}

// browserAt is the browser at path; a flatpak's launcher becomes the
// flatpak, so that the profile directory is granted to it.
func browserAt(path string) browser {
	if id, ok := program.FlatpakApp(path); ok {
		return browser{ref: id, flatpak: true, found: true}
	}
	return browser{ref: path, found: true}
}

var (
	// ErrNotExecutable means the path is not a program that can be run.
	ErrNotExecutable = errors.New("not an executable file")
	// ErrNotChromium means the program is not a Chromium-based browser.
	ErrNotChromium = errors.New("not a Chromium-based browser")
)

// chromiumVersion is the four-part version every Chromium-based browser
// prints: Chromium, Chrome, Brave, Edge. Firefox prints two or three parts.
var chromiumVersion = regexp.MustCompile(`(^|\s)\d+\.\d+\.\d+\.\d+(\s|$)`)

// CheckBrowser makes sure the program at path is a Chromium-based browser,
// which Mini Apps are driven in over the DevTools protocol, and returns what
// it is, such as "Chromium 152.0.7977.82".
//
// It runs the program with --version, so it is for a path the user picked
// and trusts to be a program, not for any file. On Windows, where a browser
// opens a window for --version, it reads the version resource instead.
//
// The answer is kept for each state of the file: the settings ask about the
// file the user picked off the frame, and a launch, or the choice of the
// player, asks again about the same file on it.
func CheckBrowser(ctx context.Context, path string) (string, error) {
	if !filepath.IsAbs(path) || !program.IsExecutable(path) {
		return "", ErrNotExecutable
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", ErrNotExecutable
	}
	key := checkKey{path, info.Size(), info.ModTime()}
	checksMu.Lock()
	result, ok := checks[key]
	checksMu.Unlock()
	if ok {
		return result.banner, result.err
	}
	banner, err := checkBrowser(ctx, path)
	// A timeout or a cancelled context says nothing about the file.
	if ctx.Err() == nil {
		checksMu.Lock()
		checks[key] = checkResult{banner, err}
		checksMu.Unlock()
	}
	return banner, err
}

// checkKey identifies a program file as it was when it was checked.
type checkKey struct {
	path    string
	size    int64
	modTime time.Time
}

type checkResult struct {
	banner string
	err    error
}

var (
	checksMu sync.Mutex
	checks   = map[checkKey]checkResult{}
)

// checkBrowser is CheckBrowser, asked every time.
func checkBrowser(ctx context.Context, path string) (string, error) {
	var banner string
	if runtime.GOOS == "windows" {
		product, version, err := program.FileVersion(ctx, path)
		if err != nil {
			return "", err
		}
		banner = product + " " + version
	} else {
		b := browserAt(path)
		name, args := b.command("")
		var err error
		banner, err = program.Banner(ctx, name, append(args, "--version")...)
		if err != nil {
			return "", err
		}
	}
	if !chromiumVersion.MatchString(banner) {
		return "", ErrNotChromium
	}
	return banner, nil
}

// findBrowser picks the browser to launch, once: the answer is asked for on
// every frame that draws the page, and finding one means running every
// candidate.
//
// A Mini App is a web page from a stranger, so the browser's engine is the
// sandbox it runs in, and the newest engine has the fewest known holes. Which
// install is newest cannot be guessed from how it was installed — a flatpak
// may be ahead of the distribution's package or behind it — so every browser
// found is asked for its version and the newest one wins.
func findBrowser() browser {
	if os.Getenv(BrowserEnv) == "" {
		if b := customBrowser(); b.found {
			return b
		}
	}
	return autoBrowser()
}

// FoundBrowser is the browser found on this system, or named by BrowserEnv,
// passing over the one SetBrowser picked: the path of a program or the id of
// a flatpak, and what it printed for --version. ref is "" when there is none.
func FoundBrowser() (ref, banner string) {
	found := autoBrowser()
	if !found.found {
		return "", ""
	}
	return found.ref, found.banner
}

// autoBrowser is findBrowser without the browser the user picked.
func autoBrowser() browser {
	browserOnce.Do(func() {
		if choice := os.Getenv(BrowserEnv); choice != "" {
			if strings.Contains(choice, ".") && flatpakInstalled(choice) {
				browserFound = probe(browser{ref: choice, flatpak: true, found: true})
			} else if path, err := exec.LookPath(choice); err == nil {
				browserFound = probe(browser{ref: path, found: true})
			}
			return
		}
		if program.Searching() {
			browserFound = newest(candidates())
		}
	})
	return browserFound
}

// candidates lists every browser installed, in order of preference, each
// asked for its version. The questions run at once: a flatpak takes about a
// second to answer.
func candidates() []browser {
	var found []browser
	seen := map[string]bool{}
	for _, name := range nativeBrowsers {
		path, err := program.LookPath(name)
		if err != nil {
			continue
		}
		// chromium and chromium-browser are often one program under two names.
		real, err := filepath.EvalSymlinks(path)
		if err != nil {
			real = path
		}
		if !seen[real] {
			seen[real] = true
			found = append(found, browser{ref: path, found: true})
		}
	}
	flatpakAt := len(found)
	found = append(found, make([]browser, len(flatpakBrowsers))...)

	var wg sync.WaitGroup
	for i := range found[:flatpakAt] {
		wg.Go(func() { found[i] = probe(found[i]) })
	}
	for i, id := range flatpakBrowsers {
		wg.Go(func() {
			if flatpakInstalled(id) {
				found[flatpakAt+i] = probe(browser{ref: id, flatpak: true, found: true})
			}
		})
	}
	wg.Wait()
	return found
}

// newest returns the browser built on the latest Chromium, the earliest in
// the list among equals. A browser that would not tell its version is chosen
// only when no other is found.
func newest(found []browser) browser {
	var best browser
	for _, b := range found {
		if !b.found {
			continue
		}
		if !best.found || b.version.newer(best.version) {
			best = b
		}
	}
	return best
}

// probe asks b for its version. On Windows, where a browser prints nothing
// for --version and opens a window instead, it reads the version resource.
func probe(b browser) browser {
	var banner string
	if runtime.GOOS == "windows" && !b.flatpak {
		product, version, err := program.FileVersion(context.Background(), b.ref)
		if err != nil {
			return b
		}
		banner = product + " " + version
	} else {
		prog, args := b.command("")
		var err error
		banner, err = program.Banner(context.Background(), prog, append(args, "--version")...)
		if err != nil {
			return b
		}
	}
	b.banner = banner
	b.version = parseVersion(b.banner)
	return b
}

func flatpakInstalled(id string) bool {
	if _, err := exec.LookPath("flatpak"); err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "flatpak", "info", id).Run() == nil
}

// version is a Chromium version, as many of its numbers as are known.
type version []int

// parseVersion reads the Chromium version from what a browser printed for
// --version: "Chromium 152.0.7977.82 for Linux Mint", "Google Chrome
// 140.0.7339.80". Brave prints its own version after the Chromium major —
// "Brave Browser 148.1.90.124" — so only its first number is Chromium's.
func parseVersion(banner string) version {
	for _, field := range strings.Fields(banner) {
		parts := strings.Split(field, ".")
		if len(parts) < 2 {
			continue
		}
		v := make(version, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.Atoi(part)
			if err != nil {
				v = nil
				break
			}
			v = append(v, n)
		}
		if v == nil {
			continue
		}
		if strings.HasPrefix(banner, "Brave") {
			v = v[:1]
		}
		return v
	}
	return nil
}

// newer reports whether v is a later Chromium than other. Only the numbers
// both know are compared, so a Brave build, which tells only its Chromium
// major, ties with every Chromium of that major. Any version is newer than an
// unknown one.
func (v version) newer(other version) bool {
	if other == nil {
		return v != nil
	}
	for i := 0; i < len(v) && i < len(other); i++ {
		if v[i] != other[i] {
			return v[i] > other[i]
		}
	}
	return false
}
