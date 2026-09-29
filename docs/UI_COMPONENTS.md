# Messenger UI components

The messenger's UI (`internal/messenger/ui`) is built from a small set of its
own components on top of Gio and a few `gio-mw` widgets. New UI reuses them,
so that the same thing looks and moves the same everywhere: a tab row is a
`tabRow`, a clickable row is a `surface`, a dialog is a `modal`. Before
writing a new one, look here; when a component is missing, make it general,
put it next to these, and add it to this page.

## Rules

- **Update, then Layout.** A component handles its input in `Update` (or a
  `Clicked` method) before the frame is drawn, and only draws in `Layout`.
  `App.Update` runs before `App.Layout`; a component that is only laid out
  handles its events at the start of its own `Layout`, as `modal` does.
- **The caller keeps the state.** Components keep what they need to draw
  and animate (clicks, tweens, scroll positions), not what is shown: the
  active tab of a `tabRow`, the section of the chat list or the text of a
  message come from the caller or the `Store`.
- **Colors and type come from the theme.** `scheme(gtx)` returns the MD3
  color scheme; text uses `token.Typestyle…`. No literal colors.
- **Text comes from the catalog.** Every string is a key of
  `localization.Catalog` (`l.T`, `l.Format`, `l.Count`), with Russian and
  English values. When Telegram Desktop shows the same text, the key maps to
  its `lng_…` key in `TelegramKeys`, and the server's language pack wins.
- **Motion follows the settings.** Animations use `wdk` tweens with `token`
  durations and easings; pass `animate` (from `Window.Motion`) where a
  component can do without, as the spoiler does.
- **Images are uploaded once per frame.** Draw decoded pictures through
  `imageOps.Op`, whose `BeginFrame`/`EndFrame` drop textures not drawn.
- **Stores change in the background.** A store calls its `changed` callback,
  which invalidates the window; components read the store again on the next
  frame and never block on it.

## Building blocks (`common.go`, `surface.go`, `buttons.go`)

| Component | Use it for |
|---|---|
| `surface` + `surfaceStyle` | Anything clickable: background of the selected state, hover and press state layer, ripple, pointer cursor, accessibility label. Rows, chips, tabs and text buttons all embed it. |
| `textButton(gtx, *surface, text)` | A text-only button in the primary color, as in dialogs and status lines. |
| `tonalButton(gtx, *surface, text, count)` | A filled button in the secondary container color with an optional count, as the actions over selected messages. |
| `navigationButton` | Icon and title with a 48 dp target, for navigation controls. |
| `button.Filled()` / `button.Text()` (gio-mw) | The main and the secondary action of a dialog or a card, side by side with a `layout.Spacer{Width: 8}`. |
| `label`, `centeredLabel` | Text in a typestyle and color, limited to a number of lines (0 is unlimited). |
| `card(gtx, content, padding)` | A rounded surface container; `defaultCardPadding` for settings and profile cards. |
| `pill(gtx, text)` | Text on a rounded plate, like service messages and the "Choose a chat" hint. |
| `toast` (`toast.go`) | Every error and outcome of an action (sent, copied, saved, failed): a grey plate at the bottom of the list or dialog it is about, gone after four seconds, a new one replacing it. Never keep such a message as text on screen. The history's is `chatPage.toast`, over the composer, floating or classic; a dialog's is `modal.Toast`, under the dialog when there is room; the chat list, `scrollPage` (settings, profile), the picker and the photo viewer have their own. A failure that repeats on every refresh is told once: compare with the one told last. A load that failed keeps a round retry button (icon only, `history_jump.go`, with the ones that scroll to the start and the end of the history), without the error text. Errors of fields with room beside them (sign-in, passwords, profile, program paths) stay under the field. |
| `vspace(dp)` | Vertical gap in a `layout.Flex`. |
| `fillRect`, `fillRounded` | Filling a size with a theme color. |
| `offset`, `inRect`, `exact` | Placing a widget at a point, in a rectangle or at an exact size. |
| `avatar` / `App.layoutAvatar` | A chat's photo, or its colored initials; pass `layoutAvatar` down as `avatarLayout`. |
| `withBadges` / `App.badges` | A name with the Premium star, emoji status, check mark or scam marks around it. |
| `drawBadge`, `drawBadgeRight` | Unread counters. |
| `flatEditor` | A borderless one-line editor with a hint, as in the composer and the picker's search. |
| `textField` | A one-line outlined field with a label, which can mask what is typed (passwords). |
| `openBrowser` | Opening a link in the system browser; call it off the frame goroutine. |

## Containers and navigation

| Component | Use it for |
|---|---|
| `tabRow` (`tabs.go`) | Tabs of equal width: centered titles, the active one in the primary color, an indicator that slides to the tab switched to, ripple on each tab. `Slide` brings in the content of the tab switched to from its side. Used by the composer's picker and the search. |
| `folderChip` (`folderbar.go`) | A capsule that is selected or not, with an optional counter: folders in the compact layout, sections of the search. Put several in a horizontal `scroll.List`. |
| `modal` (`modal.go`) | A dialog over a scrim: animates in and out, closes on Escape or a click beside it, takes the focus. The owner keeps what it shows until `Layout` reports it closed. Examples: `sessionEndedDialog`, `frozenView`, the delete dialog. |
| `contextMenu` (`contextmenu.go`) | A menu that grows from a corner of a rectangle, such as the attachment menu. |
| `scrollPage` (`pages.go`) | A centered scrollable column for a page, such as the profile. |
| `settingsItem`, `settingsChoiceCard` (`settings.go`) | A settings row with an icon, title and subtitle; a card with a title, choices and a hint. |
| `buttonRow` (`security.go`) | A secondary action at the start of a row and the main one at its end. |
| `navButton`, `sidebar` | The sidebar's sections and folders. |
| `splitter` | The draggable edge between the chat list and the page. |
| `spoiler` | Covering a value, such as a phone number in visual privacy mode, until it is clicked. |
| `loadingIndicator` | The circular progress indicator, labeled for accessibility: `page` fills a page, `sized` draws it d wide, `centered` in the middle of a line. Each place that loads keeps its own. |
| `scroll.List`, `search.Bar` (gio-mw) | Scrolling lists with a scrollbar; the search field of the chat list. |
| `toggle`, `checkbox`, `radio`, `slider` (gio-mw) | Settings controls. |

## Messages (`history_bubble.go`)

The history is drawn as materialgram draws it:

- `messageJoins` groups a sender's messages in a row (same sender, same
  day, less than 15 minutes apart): the group shows the sender's name at
  its top and one avatar, which sticks to the bottom of the view while the
  group is in it (`stickyAvatars`).
- `bubbleShape`/`shapeOf` round a bubble 16 dp, and 6 dp where it touches
  its group; `chatThemeController.Bubble` takes the same shape.
- `senderColor` colors a sender's name from its avatar color, readable on
  both themes; `datePill` is the tinted date over the history.
- `replyQuote` is the quote of a replied message, which jumps to it;
  `messageFooter` has the comments, views, author, edit mark and time;
  `reactions` wraps reaction chips.
- Channel posts have no avatar; groups, channels and bots have an icon of
  their kind before the title in the chat list (`chatKindIcon`).

## Messenger views worth copying

- **Chat list rows** (`chatlist.go`): `layoutRowWith` draws any row that looks
  like a chat (avatar, title, time, line, counter) with a `surface` of the
  caller's; search results reuse it for found messages.
- **Devices** (`sessions.go`): the settings section of the account's
  sessions, grouped as Telegram Desktop's Active Sessions, with a dialog of a
  session's details; it asks again every minute while it is open.
- **Search** (`search.go`): a `tabRow` over `folderChip` capsules over a list
  of results with a status line and its action button: a pattern for any
  panel with modes and filters. `recentSearch` (`search_recent.go`) is its
  history while nothing is typed, with a menu on a right click that takes
  the keyboard for its Escape (`key.FocusFilter`, then `key.FocusCmd`).
- **Box for sending files** (`composer_send_files.go`): a `modal` of what was
  chosen in the attachment menu, as Telegram Desktop's: media as a grid of
  squares (one picture alone keeps its shape), other files and, with "Send as
  documents", everything as rows with a thumbnail, a caption editor that takes
  the composer's text and Enter to send, and the checkboxes that
  `sendfiles` says apply (`HasGroupOption`, `HasDocumentsOption`,
  `HasHighQualityOption`). Files are looked at in the background
  (`sendfiles.Inspect`, `Thumbnail`; the first frame of a video from
  FFmpeg, when there is one); a file that cannot be sent is told in the
  dialog's toast. What it sends is `model.OutgoingFiles` in an
  `OutgoingMessage`, so retries reuse the composer's identity handling.
- **Bots** (`bot.go`): `botPage` on the chat page. The buttons under a message
  are `textButton`s in the bubble (`history_bubble.go`); a callback asks the store
  (`model.BotStore`) in a goroutine and comes back through `botUpdate` to the
  toast. The reply keyboard is a strip under the composer's bar, over which the
  reply strip stacks: `keyboardHeight` is added to `replyHeight` wherever the
  history reserves room for the composer, and `composer.Layout` draws it.
  An empty chat with a bot draws `layoutStart` in the bar's place. Typing
  `/` lists commands over the composer (`layoutCommands`).
- **Menu of a chat in the list** (`chatlist_menu.go`): a right click on a row
  opens it at the pointer, as a `contextMenu` in the header menu's style. The
  rows and the list take the presses as `search_recent.go`'s do (`PassOp` areas
  over what is clicked). Items are `chatRowPin`/`chatRowUnpin` (only in the list
  of all chats) and `chatRowRead`; a pin over the limit is answered by the
  list's `toast` without asking the store, and a store's failure comes back to
  the toast on the next frame. A pinned chat with nothing unread draws
  `drawPin` where its counter would be.
- **Forum page** (`forum.go`): the topics of a forum chat in place of its
  history, under the chat's header (which opens its info). A row is drawn as the
  chat list's: `surface`, icon plate (`fillRounded`, the topic's colour, a house
  for General, `drawPin`/lock for the marks), badges through `drawBadgeRight`.
  A topic opens as `commentsView{topic: true}` shown by `layoutComments` on
  the thread's `chatPage`, whose `topic` flag makes it read what it shows.
  `FORUM_PNG` saves the list.
- **Message sent** (`history_send.go`): a message the composer sent, when it
  shows at the end of a history that is at its end, flies up from the
  composer to its place while it fades in (300 ms); the rest of the history
  moves at once. The composer notes each send (`noteSent`) and the page takes
  a note for each message of the user that comes (an outgoing one, or any in
  Saved Messages, where Telegram does not mark them outgoing), so one from
  another device does not fly. With the animations off, by the setting or the window, or
  when the history is not at its end, the message is in its place at once.
- **Composer picker** (`composer_picker.go`): installed sticker/emoji packs and
  a Trending section for server-featured packs. A featured sticker is sent
  like an installed one; the pack's title row opens the sticker set dialog.
  The emoji tab has Telegram Desktop's seven static sections (people, nature,
  food, activity, travel, objects, symbols and flags), each with a title, and a
  button in the footer for each that takes the list to it; the button of the
  section in view is lit. Their emoji are in `emoji_sections.go`, and their
  names and keywords, in English and Russian, in `emoji_keywords.go`: both are
  generated by `go run ./cmd/emoji-sections -unicode emoji-test.txt -cldr cldr/common`
  from Unicode's `emoji-test.txt` (https://unicode.org/Public/emoji/16.0/) and
  CLDR's annotations (`common/annotations` and `annotationsDerived` of
  release-46, https://github.com/unicode-org/cldr). They are shown once the
  tab's page has come, not while it loads: the recent emoji arrive with it,
  above them, and only the first page is waited for (`pageLoaded`).
  The search of the emoji tab (`emoji_search.go`) is done by the client, at
  once, offline: by the words of the names and keywords in the language and in
  English, an exact name or keyword first; Telegram's own results, custom
  emoji, come after them. What a search found stays until the next one has
  come, and it keeps the recent and the packs of the page, so that the list
  does not empty and fill as a query is typed or cleared.
- **Frozen account** (`frozen.go`): one view shared by a bar over the chat
  list, a bar in place of the composer and a `modal`.
- **Sticker set** (`sticker_set.go`): a `modal` with an animated media grid and
  add/remove action, shared by sticker and custom emoji packs. The header's
  more menu exports the original documents as a ZIP archive. A click sends the
  sticker or puts the emoji into the draft (`messageComposer.chooseIn`), added
  set or not; emoji cells are 40 dp, sticker cells 64 dp. Stickers animate on hover even
  when automatic animations are off (also in chat and the composer picker).
  In the set's grid and the picker a sticker starts once the pointer has
  rested on it for 100 ms and no list has scrolled for 250 ms (`hoverPlay`),
  so sweeps and scrolls start no decoder. What is cheap to play
  (`chatmedia.Manager.CheapToPlay`: Lottie, drawn by the in-process WASM
  runtime, and a video sticker whose loop is decoded already) plays at once,
  as does anything in chat.
  A set opened again shows at once as last fetched (`model.StickerSetCache`),
  with its stickers' first frames, and is fetched again behind it; offline,
  it stays as cached. Loading stays at the center of the dialog. The creator menu action opens
  an accessible user or copies the creator ID hint, decoded from known pack
  ID formats; this is undocumented Telegram metadata, not verified authorship.
- **Message menu** (`history_menu.go`): a right click on a message opens a
  `contextMenu` at the pointer with Telegram Desktop's actions in its order.
  The area that takes right clicks passes every input on, and a closing menu
  lets clicks through. Actions reuse the selection's forward and delete
  dialogs (`openForwardParts`, `openDeleteParts`, `rightsFor`). The emoji
  packs item lists several sets in `emojiPacksDialog` (`emoji_packs.go`).
- **External player** (`player.go`): media opens through `chatPage.play`,
  which asks once in a `modal` when both mpv and VLC are installed and none
  is chosen; `playerSettings` is the same choice in the settings.
  `programSetting` (`programs.go`) is a card for a program the app runs — the
  one found, or a file the user picks, kept only after it passed a check. Use
  it for any new external program. FFmpeg uses the same picker and validation;
  its path also applies to GIFs and animated avatars (with a companion
  `ffprobe` beside it, or on PATH). The internal WebM sticker player defaults
  to FFmpeg when found and falls back to WASM on decode failures; WASM can
  also be selected explicitly. Static WebM previews always decode one frame
  through WASM, then close the decoder instances. FFmpeg only starts for
  playback; dormant previews wait for an event without a frame timer. The
  choice does not affect Lottie, GIFs or avatars. Stickers do not require
  `ffprobe`. The internal MP4 animation player (`decoderSettings.animations`,
  `sticker_player.go`) is the same choice for GIFs and animated avatars:
  FFmpeg when found, otherwise, or when chosen, FFmpeg's H.264 decoder in a
  WASM sandbox, one GIF at a time on hover. That decoder is not in the
  binary: `wasmmodule.AVCDec` downloads it the first time it is needed
  from [libavcodec-wasm](https://github.com/komarugif/libavcodec-wasm), at a
  pinned commit, checks its SHA-256 and keeps it in the cache directory;
  `KOMARUGRAM_AVCDEC` names a build of the user's instead (a path or a URL).
  Every FFmpeg the client runs, voice messages' and the `ffprobe` beside it
  for videos sent as media included, is `video.ResolveFFmpeg` of the path in
  the settings, else PATH. Without one, the microphone (`composer_voice.go`)
  opens the system's file chooser for Opus in OGG, MP3 or M4A
  (`voice.FileExtensions`) and sends the file as a voice message, its
  duration read from its headers (`voice.FileDuration`), without a waveform;
  the file is the user's and is not removed. `program.LookPath` and
  `FindFlatpak` are the only searches for external programs:
  `-no-integrations` turns them off. Voice messages and music play in the
  client (`audio_player.go`): `audioPlayer`, one per chat page;
  `voiceLayout`, the row with its button, waveform (`drawWaveform`) and
  time, and `musicLayout`, with the title, the performer and a bar
  (`drawBar`); `audioRow.update` takes their clicks and drags. `showWaveform`
  works one out, as soon as it is shown, for a voice message sent without
  one; `audioFormat` picks libopus, dr_libs or fdk-aac, and a format none
  takes goes to the external player (`takeExternal`); tests give
  `audioPlayer.play` a silent playback, so that nothing is heard. A GIF shown still starts no process either: it shows Telegram's
  thumbnail, fetched before the file (a real GIF file's first frame is
  decoded in Go); large GIFs have none and stay blank until played, as in the
  official clients. `ffprobe` and `ffmpeg` start when it plays, and `ffmpeg`
  exits a second after it stops.
- **Without a bubble** (`history_unwrapped.go`): stickers and a lone emoji
  (half a sticker's size) lie on the chat's background, as in the official
  clients; `infoPlate` is their time, and a reply or forward goes in a card
  at the side.
- **Reactions** (`reactions.go`): chips under a message toggle a reaction
  (`model.Reactor`); `reactionStrip` is the row at the top of the message
  menu that expands into all the chat's reactions. A double click on a
  bubble, off its text, puts the default reaction (`model.QuickReactor`);
  `reactedDialog` (`reacted.go`) lists who reacted, a tab for each reaction.
- **Comments** (`comments.go`): `commentsBar` ends a channel post's bubble,
  clipped to its shape; the comments open in a second `chatPage` under a
  `chatHead` — `layoutChatPageHead` with a back button in place of the
  avatar — over their channel.
- **Service messages** (`history_service.go`): an action, as a user added or
  a message pinned, is its words on a plate across the middle of the
  history (`servicePill`), worded in the UI's language by
  `localization.Catalog.Service` from `model.ServiceAction`. A pin's plate
  quotes the message and shows it on a click; a call is a bubble.
- **Header actions** (`chat_menu.go`, `chat_search.go`): the search and
  menu buttons at the end of a chat's header. The search is a field over
  the header (`model.ChatSearcher`) with a counter and buttons to the older
  and newer found messages; the one shown is tinted for a moment
  (`highlight`). The menu is a `contextMenu` under its button.
- **Message shot** (`snapshot.go`): the dialog the selection's snapshot
  button opens, as AyuGram's message shot box: a preview rendered off the
  frame (`buildSnapshot`, `renderSnapshot`), the theme and what it shows,
  and buttons that save the PNG or copy it (`clipboard.WriteCmd` with
  `image/png`).
- **Pinned bar** (`pinned_bar.go`): under the chat's header, the latest
  pinned message above the history's bottom (`model.PinnedSource`); a click
  goes to it and shows the one before, as Telegram Desktop's does. Its line
  has a segment for each pinned message, up to four.
- **Reply strip** (`composer_reply.go`): the message a draft replies to, over
  the composer; the history's end and what opens from the composer move up by
  its height (`replyHeight`).

## Don'ts

- Don't bring in another `gio-mw` widget for something listed here, such as
  `tab.Group` for tabs: it looks and moves differently from the rest.
- Don't show loading as text ("Loading…", "Searching…"): use a
  `loadingIndicator` where the content will appear. Text is for what went
  wrong or what the user can do.
- Don't use `widget.Clickable` alone for a visible element: without a
  `surface` it has no hover, press or ripple.
- Don't keep a copy of store data in a component beyond a frame, except
  what an animation needs to finish, as `modal` does.

## Checking UI changes

Look at the result, not only at the tests: alignment, centering and how the
new piece sits next to its neighbours.

- **Render tests.** `renderFrames` in `session_ended_render_test.go` draws a
  component headlessly for enough frames to finish its animations and saves a
  PNG; `TestRenderComposer`, `TestRenderSettingsAccounts` and
  `TestRenderSessionEnded` show how. `TestRenderStickerSet` saves both sticker
  and emoji pack states; `TestRenderMessageMenu`, the message menu over a
  reply, with a floating and a classic composer. They are skipped unless their
  variable (`COMPOSER_PNG`, `COMPOSER_MOTION_PNG`, `SETTINGS_PNG`,
  `ACCOUNTS_PNG_DIR`, `SESSION_PNG_DIR`, `STICKER_SET_PNG_DIR`, `MENU_PNG`, `SAVED_EMPTY_PNG`, `VIEWER_PNG`, `PLAYER_PNG`, `COMMENTS_PNG`, `UNWRAPPED_PNG`, `SERVICE_PNG`, `JUMP_PNG_DIR`, `PINNED_PNG`, `REACTED_PNG`, `CHAT_SEARCH_PNG`, `SHOT_PNG`, `SESSIONS_PNG`, `FORUM_PNG`, `SHARED_PNG`, `TOAST_PNG_DIR`, `AUDIO_PNG_DIR`) is set; `go run ./cmd/render-all` runs them all. `TOAST_PNG_DIR` gets a toast in every place that has one, light and dark. The last one shows Saved Messages before its
  first dialog exists. `COMPOSER_VIEW=emoji-search`, the picker's emoji found for a word;
  `COMPOSER_VIEW=featured-stickers` or `featured-emoji`
  with `COMPOSER_PNG` shows the picker's recommendations; `COMPOSER_VIEW=voice`,
  a voice message being recorded; `files`, `files-one`, `files-documents`,
  `files-many` and `files-caption`, the box for sending files.
  `SETTINGS_SECTION=premium` with `SETTINGS_PNG` shows the Premium section of
  an account without Premium and with Local Premium on (`LOCAL_PREMIUM=off`,
  off).
  `SETTINGS_SECTION=integrations` with `SETTINGS_PNG` shows the choice of the
  external player; `PLAYER_PNG`, the dialog that asks for it. `MENU_PNG` also
  saves the menu with reactions (`-reactions*.png`); `COMMENTS_PNG`, the
  header of the comments page; `UNWRAPPED_PNG`, stickers and lone emoji; `SERVICE_PNG`, service messages
  and calls.
- **The app itself.** `go run ./cmd/messenger -demo` runs without an account.
  On Linux under X11, `xdotool` drives a window (`mousemove --window … click`)
  and `xfce4-screenshooter -w` saves the active one.
