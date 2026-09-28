# AGENTS.md

Start here. This file is for AI agents working in this repository: what to
read, what to run, and what not to break. `README.md` is the human
documentation of the same things at length; read its sections only when a
task needs them.

## What this is

A Telegram desktop client in Go on Gio (module `komarugram`), a replacement for
Telegram Desktop (tdesktop) that works offline from its local cache, plus
`kitchen`, a demo of widgets. The maintainer writes in Russian: answer in
Russian; code, comments and docs are in English.

| Path | What |
|---|---|
| `cmd/messenger` | Entry point; `account_host.go` runs accounts, windows and their stores |
| `internal/messenger/ui` | All messenger UI. Components: `docs/UI_COMPONENTS.md` |
| `internal/messenger/tgstore` | `model.Store` backed by Telegram (gotd): chats, history, search, updates |
| `internal/messenger/historycache` | Per-account SQLite cache: messages (JSON), media, layouts, FTS5 search index |
| `internal/messenger/account` | Registry, sign-in, tdata import, one client per auth key |
| `internal/messenger/model` | Data types and the interfaces the UI reads |
| `internal/messenger/{security,securedb}` | TPM-sealed key, encrypted SQLite (Adiantum VFS) |
| `internal/messenger/mockstore` | Demo store for `-demo` |
| `pkg/*` | Reusable parts: `dcpool` (download connections), `tdata`, media decoders |
| `third_party/gio`, `third_party/gio-mw` | Forks, wired with `replace`; changes to Gio go in `third_party/gio/LOCAL_CHANGES.md` |
| `ayugram`, `materialgram` | Sources of two Telegram Desktop forks, without git history: the reference for Telegram behavior, texts and UI. AyuGram (AyuGram/AyuGramDesktop db3b989, 2026-08-08) adds features such as message snapshots (`ayu/features/message_shot`); materialgram (kukuruzka165/materialgram d31cdec, 2026-07-25) restyles the UI. Only the submodules `Telegram/lib_ui`, `lib_base` and `lib_tl` were fetched; the rest of `lib_*` and `ThirdParty` are empty |

## Read before

- What to do next: `docs/PLAN.md` — porting features of AyuGram and materialgram,
  with the rules for it (behavior, not code: they are GPLv3), and the
  technical debts.
- UI work: `docs/UI_COMPONENTS.md`. Reuse `surface`, `tabRow`, `folderChip`,
  `modal`… instead of new widgets; check the result by looking at it.
- Telegram behavior (errors, limits, flows): how the forks do it in
  `ayugram/Telegram/SourceFiles` and `materialgram/Telegram/SourceFiles`
  (upstream tdesktop code plus their own), widgets in `*/Telegram/lib_ui`,
  and https://core.telegram.org/api. Look at both forks: a feature may be
  one fork's own, as AyuGram's snapshots are.
  Texts shown to the user map to tdesktop's `lng_…` keys in
  `internal/messenger/localization` (`TelegramKeys`).
- Performance or memory: `docs/PROFILING.md`; README "Pitfalls".

## Commands

```sh
go build -o /dev/null ./cmd/messenger          # never leave binaries in the tree
go vet ./internal/... ./cmd/... ./pkg/...
go test ./internal/... ./cmd/... ./pkg/...
gofmt -l internal cmd pkg
go vet gioui.org/app                           # the Gio fork, by its import path
go run ./cmd/messenger -demo                   # no account needed
```

These cross-builds must keep working:

```sh
go build -tags nowayland ./cmd/messenger       # also nox11, novulkan
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o /dev/null ./cmd/messenger ./cmd/kitchen
GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -o /dev/null ./cmd/messenger
```

Render tests save PNGs when their variable is set, and are skipped
otherwise: `COMPOSER_PNG`, `SETTINGS_PNG`, `SESSION_PNG_DIR`, `STICKER_SET_PNG_DIR`, `MENU_PNG`, `SAVED_EMPTY_PNG`, `VIEWER_PNG`, `PLAYER_PNG`, `COMMENTS_PNG`, `UNWRAPPED_PNG`, `SERVICE_PNG`, `PINNED_PNG`, `REACTED_PNG`, `CHAT_SEARCH_PNG`, `SHOT_PNG`, `SESSIONS_PNG`
(see `docs/UI_COMPONENTS.md`). `pkg/h264`'s test needs `KOMARUGRAM_AVCDEC`
set to the absolute path of an `avcdec.wasm`, which is not in this repository
([libavcodec-wasm](https://github.com/komarugif/libavcodec-wasm)).

## Rules

- **Root causes, not workarounds.** Measure before and after; say what was
  verified and what was not.
- **Live-test only the operating system the maintainer is using.** Other operating systems are tested manually by the maintainer: say when a change needs their check.
- **Memory is a feature.** The app must give memory back to the OS (hidden
  windows drop their GPU context, `malloc_trim` after a window closes). When
  RSS grows but the Go heap does not, look at the native side.
- **Binary size.** Compare stripped builds (`-ldflags='-s -w'`) after adding
  a dependency. Nothing may call `reflect.Type.Method`/`MethodByName` (godbus
  `Export` did, and doubled the binary through gotd's `tg`).
- **One connection per auth key, main sessions only on the main DC.** Two
  clients on one key, or parallel sessions to the main DC's non-media
  addresses, make Telegram kill the key with `AUTH_KEY_DUPLICATED`. `dcpool`
  sends downloads from the home DC to its media-only addresses for this.
- **Accounts are the maintainer's real ones.** Read-only calls are fine for
  testing; never send, edit or delete. Don't search public posts: each search
  spends one of the day's free searches.
- **Found messages are not saved to history.** The cache keeps each chat's
  history without gaps; single messages from search would break that.
- **SQLite payloads are JSON BLOBs:** `json_extract(CAST(payload AS TEXT), …)`.
  FTS5 is registered for every connection in `historycache` (`AutoExtension`);
  its triggers need it.
- **A new test must fail without the fix.** Break the code, run, restore.

## Gio in short

- Handle input at the top of `Layout` (or in `Update`), before drawing what it
  changes; otherwise nothing redraws until the next event.
- Zero `gtx.Constraints.Min` before content that should take its own size.
- Each window has its own goroutine and GPU context; stores are shared and
  must be safe for concurrent use.
- On X11, `Window.Run`/`Option` run on the caller's goroutine: anything a
  driver method uses must be set before `SetDriver`.

The full list, with the story behind each item, is README "Pitfalls met while
working on this code".

## Testing the app live (Linux, X11/XFCE)

- Run the built binary from a scratch directory in the background. Stop it
  with `pkill -x messenger`: `pkill -f <path>` also matches, and kills, the
  shell that runs it.
- Drive a window with `xdotool mousemove --window <id> x y click 1` (client
  coordinates) and `xdotool type`; find it with `wmctrl -l`; screenshot the
  active window with `xfce4-screenshooter -w -s file.png`.
- Close windows with `wmctrl -i -c <id>`, not `xdotool windowclose`: Gio never
  sees a `DestroyEvent` then.
- The tray (SNI) can be called over D-Bus: `busctl --user call
  org.kde.StatusNotifierItem-<pid>-1 /StatusNotifierItem
  org.kde.StatusNotifierItem Activate ii 0 0`.
- The instance socket is `$XDG_RUNTIME_DIR/komarugram-go.sock`; a path over 107
  bytes fails, so point `XDG_RUNTIME_DIR` at a short symlink when needed.
- Local data: `~/.config/komarugram-go` (accounts, sessions, `history.db.*`),
  crash reports in `~/.cache/komarugram-go/crashes`.
