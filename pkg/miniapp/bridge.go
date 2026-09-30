// SPDX-License-Identifier: Unlicense OR MIT

// Package miniapp runs a Telegram Mini App in the user's own Chromium and
// speaks to it over the Chrome DevTools Protocol.
//
// Official clients embed a webview and inject an object into it. There is no
// webview here and no intention of shipping one, but the interface a Mini App
// expects is small enough to provide from the outside: the page reaches the
// client through TelegramWebviewProxy.postEvent, the client reaches the page
// through Telegram.WebView.receiveEvent, and the launch parameters arrive in
// the URL fragment. CDP can do all three — a binding for the first, an
// evaluation for the second, a navigation for the third.
//
// The browser runs as a separate process with its own profile, so a Mini App
// never sees the user's cookies and cannot touch this process's memory.
package miniapp

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// bindingName is the CDP binding the injected shim calls. It is deliberately
// obscure: the page should reach the client through the documented proxy, not
// by guessing this name.
const bindingName = "__tgWebviewProxyPostEvent"

// shim runs before any script of the page, which is when the Mini App SDK
// looks for its transport.
const shim = `
window.TelegramWebviewProxy = {
  postEvent: function (eventType, eventData) {
    ` + bindingName + `(JSON.stringify({eventType: eventType, eventData: eventData}));
  }
};
`

// Event is one message from the Mini App.
type Event struct {
	Type string
	Data string
	At   time.Time
}

// Params are the launch parameters a client puts in the URL fragment.
type Params struct {
	InitData    string // as returned by messages.requestWebView
	Version     string
	Platform    string
	ThemeParams string // JSON
	// Extra is more of the fragment, as a client got it from Telegram and
	// escaped as it was: launch fields this package does not know, such as
	// tgWebAppStartParam.
	Extra string
}

// Storage decides what a Mini App leaves behind and who may read it.
type Storage int

const (
	// Ephemeral gives every launch a profile of its own, removed on Close.
	// Nothing survives the window being closed.
	Ephemeral Storage = iota
	// PerApp keeps one profile per Mini App: what an app stores it finds again
	// next time, and no other app can reach it.
	PerApp
	// Shared keeps one profile for every Mini App of an account, which is what
	// an official client does.
	Shared
)

// Profile says where a Mini App's browsing data lives.
//
// A Mini App is an ordinary web application: it writes cookies and localStorage
// and expects them back. Official clients let it. Telegram Desktop hands every
// bot webview the same storage id (resolveStorageIdBots), keeps it in the
// account directory as wvbots and clears it only on logout; Telegram for
// Android runs its webview with DOM storage, databases and third-party cookies
// enabled and flushes the cookie jar to disk. So Shared is what an official
// client does, PerApp is stricter than any of them — one bot cannot read what
// another left — and Ephemeral, the zero value, keeps nothing at all.
type Profile struct {
	Storage Storage

	// Root is where persistent profiles are kept. Empty means a directory under
	// the user's cache directory.
	Root string

	// App identifies the Mini App — in a real client, the bot it belongs to.
	// PerApp needs it; the other modes ignore it.
	App string

	// Account identifies whose apps share a profile under Shared. Empty names a
	// single unnamed account, which is enough while there is only one.
	Account string
}

// resolve returns the directory to launch the browser on and whether that
// directory is this launch's alone.
func (p Profile) resolve() (dir string, ephemeral bool, err error) {
	if p.Storage == Ephemeral {
		dir, err = os.MkdirTemp("", "kitchen-miniapp-")
		return dir, true, err
	}
	root := p.Root
	if root == "" {
		cache, err := os.UserCacheDir()
		if err != nil {
			return "", false, err
		}
		root = filepath.Join(cache, "gio-kitchen", "miniapp")
	}
	switch p.Storage {
	case PerApp:
		if p.App == "" {
			return "", false, fmt.Errorf("a per-app profile needs an app to name it after")
		}
		dir = filepath.Join(root, "apps", profileKey(p.App))
	case Shared:
		dir = filepath.Join(root, "shared", profileKey(p.Account))
	default:
		return "", false, fmt.Errorf("unknown storage mode %d", p.Storage)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", false, err
	}
	return dir, false, nil
}

// profileKey turns an identifier into one directory name. Bot and account ids
// are plain numbers, but nothing here guarantees that, so anything else is
// hashed rather than flattened into a name that another key could also produce
// — or into one that climbs out of the root.
func profileKey(id string) string {
	if id == "" {
		return "default"
	}
	safe := id != "." && id != ".."
	for _, r := range id {
		if !(r == '-' || r == '_' || r == '.' ||
			(r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			safe = false
			break
		}
	}
	if safe {
		return id
	}
	sum := sha256.Sum256([]byte(id))
	return hex.EncodeToString(sum[:8])
}

// Bridge is one running Mini App: a browser process, a CDP connection and the
// two halves of the Telegram transport.
type Bridge struct {
	chrome    *exec.Cmd
	profile   string
	ephemeral bool
	exited    chan struct{}
	conn      *websocket.Conn
	browserWS string
	cancel    context.CancelFunc

	mu      sync.Mutex
	nextID  int
	replies map[int]chan json.RawMessage
	events  []Event
	err     error
	closed  bool
}

// Open starts the browser on url and installs the bridge. profile decides what
// the Mini App is allowed to keep between launches; its zero value keeps
// nothing.
func Open(ctx context.Context, url string, params Params, profile Profile) (*Bridge, error) {
	// The fragment carries the launch parameters, exactly as a webview would
	// receive them.
	return launch(ctx, url+"#"+params.fragment(), profile, Page{Width: 420, Height: 720}, true)
}

// Page is how OpenPage shows a page of the client's own.
type Page struct {
	// Width and Height are the size of the window, in the browser's pixels.
	Width, Height int
	// Args are more switches for the browser.
	Args []string
}

// OpenPage starts the browser on url in a window of its own, on a throwaway
// profile, without the Telegram transport: for a page the client itself
// shows in the browser, such as a video player. Eval reaches the page.
func OpenPage(ctx context.Context, url string, page Page) (*Bridge, error) {
	return launch(ctx, url, Profile{}, page, false)
}

// launch starts the browser on url with profile and attaches to the page,
// installing the Telegram transport when telegram is set.
func launch(ctx context.Context, url string, profile Profile, page Page, telegram bool) (*Bridge, error) {
	dir, ephemeral, err := profile.resolve()
	if err != nil {
		return nil, err
	}
	chosen := findBrowser()
	if !chosen.found {
		if choice := os.Getenv(BrowserEnv); choice != "" {
			return nil, fmt.Errorf("%s names %q, which is neither a program nor an installed flatpak",
				BrowserEnv, choice)
		}
		return nil, fmt.Errorf("no Chromium-based browser found")
	}
	discard := func() {
		if ephemeral {
			os.RemoveAll(dir)
		}
	}

	if err := seedProfile(dir); err != nil {
		discard()
		return nil, err
	}
	if err := checkNotRunning(dir); err != nil {
		discard()
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	bridge := &Bridge{
		profile:   dir,
		ephemeral: ephemeral,
		exited:    make(chan struct{}),
		cancel:    cancel,
		replies:   map[int]chan json.RawMessage{},
	}

	prog, args := chosen.command(dir)
	args = append(args,
		"--app="+url,
		"--user-data-dir="+dir,
		"--remote-debugging-port=0",
		"--no-first-run", "--no-default-browser-check",
		fmt.Sprintf("--window-size=%d,%d", page.Width, page.Height),
	)
	bridge.chrome = exec.Command(prog, append(args, page.Args...)...)
	if err := bridge.chrome.Start(); err != nil {
		cancel()
		discard()
		return nil, fmt.Errorf("start browser: %w", err)
	}
	// One waiter owns the process: Close and the launch both need to know when
	// it is gone, and only one of them may call Wait.
	go func() {
		_ = bridge.chrome.Wait()
		close(bridge.exited)
	}()

	if err := bridge.connect(ctx, telegram); err != nil {
		bridge.Close()
		return nil, err
	}
	return bridge, nil
}

// seedProfile writes the preferences the browser is to start with. A Mini App
// is a client surface rather than a page the user browsed to, so Chromium's
// offer to translate it — a bar across the top of someone else's app — has no
// place here. There is no command-line switch for it: --disable-features has
// no Translate feature to switch off, and on distributions whose launcher is a
// wrapper script it would also override the flags that script passes. The
// preference is read out of a fresh profile on every launch, which makes it
// permanent for as long as this is the only way a Mini App opens.
func seedProfile(profile string) error {
	dir := filepath.Join(profile, "Default")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	// Two files, because the browsers keep these two settings at different
	// levels: the offer to translate belongs to the profile, and Brave's notice
	// about its analytics belongs to the browser. A browser that has never
	// heard of the other one's setting ignores it.
	for path, value := range map[string]string{
		filepath.Join(dir, "Preferences"): `{"translate":{"enabled":false}}`,
		filepath.Join(profile, "Local State"): `{"brave":{"p3a":` +
			`{"enabled":false,"notice_acknowledged":true}}}`,
	} {
		// Only where there is none: the browser keeps its whole state in these
		// files, and a kept profile is handed back the ones it wrote last time.
		// The settings survive, because the browser rewrites them with it.
		if _, err := os.Stat(path); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return err
		}
		if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
			return fmt.Errorf("seed browser profile: %w", err)
		}
	}
	return nil
}

// checkNotRunning reports whether a browser already holds this profile.
// Chromium allows one process per profile: a second launch hands its window to
// the first and exits, leaving two pages on one debugging endpoint and this
// bridge attached to whichever it found first. A throwaway profile cannot
// collide, a persistent one can, so it is caught here while it can still be
// explained.
func checkNotRunning(profile string) error {
	path := filepath.Join(profile, "DevToolsActivePort")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	port, _, _ := strings.Cut(string(data), "\n")
	client := http.Client{Timeout: time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/json/version")
	if err == nil {
		resp.Body.Close()
		return fmt.Errorf("this Mini App is already open in another window")
	}
	// Nobody answered, so the file is what an earlier run left behind. It has
	// to go, or the launch below would take it for its own port.
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (p Params) fragment() string {
	fields := []string{}
	add := func(key, value string) {
		if value != "" {
			fields = append(fields, key+"="+value)
		}
	}
	// InitData is itself a query string, so the caller passes it already
	// escaped; the rest are plain tokens.
	add("tgWebAppData", p.InitData)
	add("tgWebAppVersion", p.Version)
	add("tgWebAppPlatform", p.Platform)
	add("tgWebAppThemeParams", p.ThemeParams)
	if p.Extra != "" {
		fields = append(fields, p.Extra)
	}
	return strings.Join(fields, "&")
}

// connect waits for the debugging endpoint and attaches to the page, and
// installs both halves of the bridge when telegram is set.
func (b *Bridge) connect(ctx context.Context, telegram bool) error {
	port, err := b.waitForPort(ctx)
	if err != nil {
		return err
	}
	target, err := pageTarget(ctx, port)
	if err != nil {
		return err
	}
	// Kept for Close: the browser endpoint is the only one that will shut the
	// browser down cleanly on request.
	b.browserWS, _ = browserTarget(ctx, port)
	conn, _, err := websocket.Dial(ctx, target, nil)
	if err != nil {
		return fmt.Errorf("attach to page: %w", err)
	}
	conn.SetReadLimit(8 << 20)
	b.conn = conn
	go b.read(ctx)

	type step struct {
		method string
		params map[string]any
	}
	steps := []step{{"Runtime.enable", nil}, {"Page.enable", nil}}
	if telegram {
		steps = append(steps,
			step{"Runtime.addBinding", map[string]any{"name": bindingName}},
			step{"Page.addScriptToEvaluateOnNewDocument", map[string]any{"source": shim}},
			// The page loaded before the shim existed, so it is loaded again
			// with the bridge in place.
			step{"Page.reload", map[string]any{"ignoreCache": true}},
		)
	}
	for _, step := range steps {
		if _, err := b.call(ctx, step.method, step.params); err != nil {
			return fmt.Errorf("%s: %w", step.method, err)
		}
	}
	return nil
}

// Send delivers an event to the Mini App. data must be a JavaScript
// expression, usually a JSON object literal.
func (b *Bridge) Send(ctx context.Context, eventType, data string) error {
	if data == "" {
		data = "null"
	}
	_, err := b.call(ctx, "Runtime.evaluate", map[string]any{
		"expression": fmt.Sprintf("window.Telegram.WebView.receiveEvent(%q, %s)", eventType, data),
	})
	return err
}

// Eval runs an expression in the Mini App and returns its value as a string.
// It is what a client uses to inspect or adjust the page directly — drawing
// the header and main button inside the page, for instance, since they cannot
// be drawn around someone else's window. A promise is waited for, and its
// value returned.
func (b *Bridge) Eval(ctx context.Context, expression string) (string, error) {
	result, err := b.call(ctx, "Runtime.evaluate", map[string]any{
		"expression":    expression,
		"returnByValue": true,
		"awaitPromise":  true,
	})
	if err != nil {
		return "", err
	}
	var wrapper struct {
		Result struct {
			Value any `json:"value"`
		} `json:"result"`
	}
	if err := json.Unmarshal(result, &wrapper); err != nil {
		return "", err
	}
	text, _ := wrapper.Result.Value.(string)
	return text, nil
}

// SetWindowBounds moves and resizes the window the page is in, in the
// screen's device-independent pixels, the ones window.screen measures. A
// window manager may keep a window from placing itself, as Wayland does; the
// size still applies.
func (b *Bridge) SetWindowBounds(ctx context.Context, left, top, width, height int) error {
	result, err := b.call(ctx, "Browser.getWindowForTarget", nil)
	if err == nil {
		err = replyError(result)
	}
	if err != nil {
		return fmt.Errorf("Browser.getWindowForTarget: %w", err)
	}
	var window struct {
		WindowID int `json:"windowId"`
	}
	if err := json.Unmarshal(result, &window); err != nil {
		return err
	}
	result, err = b.call(ctx, "Browser.setWindowBounds", map[string]any{
		"windowId": window.WindowID,
		"bounds": map[string]any{
			"left": left, "top": top, "width": width, "height": height,
			"windowState": "normal",
		},
	})
	if err == nil {
		err = replyError(result)
	}
	if err != nil {
		return fmt.Errorf("Browser.setWindowBounds: %w", err)
	}
	return nil
}

// replyError is the error the browser answered a call with, as read turns
// it into a result.
func replyError(result json.RawMessage) error {
	var reply struct {
		Error *string `json:"error"`
	}
	if json.Unmarshal(result, &reply) == nil && reply.Error != nil {
		return errors.New(*reply.Error)
	}
	return nil
}

// Events returns everything the Mini App has sent so far.
func (b *Bridge) Events() []Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]Event(nil), b.events...)
}

// Err reports why the bridge stopped working, if it did.
func (b *Bridge) Err() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.err
}

// Running reports whether the browser window is still open.
func (b *Bridge) Running() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return false
	}
	select {
	case <-b.exited:
		return false
	default:
		return true
	}
}

// Close stops the browser. A throwaway profile goes with it; a persistent one
// stays, holding whatever the Mini App stored.
func (b *Bridge) Close() {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	b.closed = true
	b.mu.Unlock()

	b.cancel()
	if b.conn != nil {
		_ = b.conn.Close(websocket.StatusNormalClosure, "")
	}
	b.stopBrowser()
	if b.ephemeral {
		_ = os.RemoveAll(b.profile)
	}
}

// stopBrowser ends the browser process, giving it every chance to put its
// storage away first.
//
// A browser killed outright never flushes what the Mini App stored, and a
// signal is no better while the browser is still starting up: it arrives before
// the handlers that would make it graceful, and the process dies where it
// stands. Browser.close is the browser's own shutdown path and is answered as
// soon as the endpoint is up, so it is what gets asked first.
func (b *Bridge) stopBrowser() {
	if b.chrome.Process == nil {
		return
	}
	if b.browserWS != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		conn, _, err := websocket.Dial(ctx, b.browserWS, nil)
		if err == nil {
			err = conn.Write(ctx, websocket.MessageText,
				[]byte(`{"id":1,"method":"Browser.close"}`))
		}
		if err == nil {
			select {
			case <-b.exited:
			case <-ctx.Done():
			}
		}
		if conn != nil {
			_ = conn.Close(websocket.StatusNormalClosure, "")
		}
		cancel()
	}
	select {
	case <-b.exited:
		return
	default:
	}
	if err := b.chrome.Process.Signal(os.Interrupt); err != nil {
		_ = b.chrome.Process.Kill()
	}
	select {
	case <-b.exited:
	case <-time.After(5 * time.Second):
		_ = b.chrome.Process.Kill()
		<-b.exited
	}
}

// read dispatches CDP replies and events.
func (b *Bridge) read(ctx context.Context) {
	for {
		_, data, err := b.conn.Read(ctx)
		if err != nil {
			b.mu.Lock()
			if !b.closed && b.err == nil {
				b.err = err
			}
			b.mu.Unlock()
			return
		}
		var message struct {
			ID     int             `json:"id"`
			Result json.RawMessage `json:"result"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
			Error  *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(data, &message) != nil {
			continue
		}
		if message.ID != 0 {
			b.mu.Lock()
			reply := b.replies[message.ID]
			delete(b.replies, message.ID)
			b.mu.Unlock()
			if reply != nil {
				if message.Error != nil {
					reply <- json.RawMessage(`{"error":` + jsonString(message.Error.Message) + `}`)
				} else {
					reply <- message.Result
				}
			}
			continue
		}
		if message.Method == "Runtime.bindingCalled" {
			b.handleBinding(message.Params)
		}
	}
}

func (b *Bridge) handleBinding(params json.RawMessage) {
	var call struct {
		Name    string `json:"name"`
		Payload string `json:"payload"`
	}
	if json.Unmarshal(params, &call) != nil || call.Name != bindingName {
		return
	}
	var message struct {
		EventType string `json:"eventType"`
		EventData string `json:"eventData"`
	}
	if json.Unmarshal([]byte(call.Payload), &message) != nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, Event{Type: message.EventType, Data: message.EventData, At: time.Now()})
}

func (b *Bridge) call(ctx context.Context, method string, params map[string]any) (json.RawMessage, error) {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil, fmt.Errorf("bridge is closed")
	}
	b.nextID++
	id := b.nextID
	reply := make(chan json.RawMessage, 1)
	b.replies[id] = reply
	b.mu.Unlock()

	request := map[string]any{"id": id, "method": method}
	if params != nil {
		request["params"] = params
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	if err := b.conn.Write(ctx, websocket.MessageText, payload); err != nil {
		return nil, err
	}
	select {
	case result := <-reply:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("%s timed out", method)
	}
}

// waitForPort waits for the browser to publish its debugging port, or for it to
// give up. A browser that exits this early has usually handed its window to
// another process holding the same profile.
func (b *Bridge) waitForPort(ctx context.Context) (string, error) {
	path := filepath.Join(b.profile, "DevToolsActivePort")
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(path); err == nil {
			if port, _, ok := strings.Cut(string(data), "\n"); ok {
				return port, nil
			}
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-b.exited:
			return "", fmt.Errorf("the browser exited without opening a debugging port; " +
				"another window may already hold this profile")
		case <-time.After(50 * time.Millisecond):
		}
	}
	return "", fmt.Errorf("the browser never opened a debugging port")
}

// browserTarget returns the WebSocket of the browser itself, as opposed to the
// page in it.
func browserTarget(ctx context.Context, port string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"http://127.0.0.1:"+port+"/json/version", nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var version struct {
		WS string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&version); err != nil {
		return "", err
	}
	return version.WS, nil
}

func pageTarget(ctx context.Context, port string) (string, error) {
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://127.0.0.1:" + port + "/json/list")
		if err == nil {
			var targets []struct {
				Type string `json:"type"`
				WS   string `json:"webSocketDebuggerUrl"`
			}
			err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&targets)
			resp.Body.Close()
			if err == nil {
				for _, target := range targets {
					if target.Type == "page" && target.WS != "" {
						return target.WS, nil
					}
				}
			}
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	return "", fmt.Errorf("the browser never opened a page")
}

func jsonString(s string) string {
	quoted, _ := json.Marshal(s)
	return string(quoted)
}
