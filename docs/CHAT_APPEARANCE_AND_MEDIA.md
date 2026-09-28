# Chat appearance and shared media

Click a chat's header to open its information dialog. Section rows use the
same `settingsItem` component as Settings. Sections with a known zero count
are hidden; a failed count remains unknown rather than being shown as zero.

## Browsing

- Photos and videos use a padded grid. Photos open the existing viewer;
  videos and audio stream to the external player (mpv or VLC). GIFs use aspect-aware rows and follow the
  application's animation preference.
- Files open through the system application after downloading to a private
  temporary directory, removed when the owning window is destroyed.
- Links use a dedicated list: Telegram's `WebPage` title and photo, plain
  message text capped at three lines, and clickable URLs. Spoilers remain
  concealed. No page, HTML metadata, favicon, or image URL is requested from
  a website to generate previews. Preview photos use Telegram file locations
  and the account's existing transport/proxy. Explicitly opening a link still
  uses the existing external-browser confirmation.
- Polls and saved messages use the message renderer inside the dialog.
- Public profile stories, gifts, and common groups have independent API
  sources. Story and gift identifiers are not inserted into chat history.
- Gifts use compact square cards with sender avatars. Collectible cards use
  their Telegram backdrop colors, symbol and number. A separate dialog shows
  the larger animation without that background, sender/date/value, collectible
  attributes and quantity, and a link to Telegram Stars terms. Anonymous
  senders stay anonymous. When no price is supplied for a collectible, the
  value is shown as an em dash rather than inferred from resale/conversion data.

`messages.search` and `messages.getSearchCounters` provide shared-message
pages and totals without loading history. Saved-message queries target Self
with `saved_peer_id`. Photo-viewer pagination also searches Telegram, with
cached photos available offline. Successful search pages retain their media
references and variants in the account cache. Collection queries use
`stories.getPinnedStories`, `payments.getSavedStarGifts`, and
`messages.getCommonChats`.

The shared-media manager keeps up to 160 entries and budgets decoded pixels
at 96 MiB, evicting inactive entries before visible ones. Hidden animations
pause immediately and may retain decoders for 15 seconds for scroll-back;
inactive decoders are also retired under count pressure. Gift presentation
textures have their own bounded retention. Hiding the window in the tray
uses the existing Release lifecycle, dropping pixels while retaining
navigation and scroll state.

## Appearance

The appearance page loads Telegram chat themes and the peer's selected
wallpaper, supports light/dark variants, outgoing bubble gradients, image
wallpapers, PNG/TGV patterns, intensity/inversion and blurred wallpapers.
It also imports the chat subset of `.tdesktop-theme` ZIP archives and plain
palettes, including aliases and bundled normal/tiled wallpapers. Imported
palettes do not replace the application's complete UI theme.

A preview can be applied locally to the account/chat or, for Telegram themes,
explicitly applied to chat participants through `messages.setChatTheme`.
Local overrides and remote appearances are stored in the account cache.
Freeform gradient interpolation is implemented independently; it is not a
pixel-exact reproduction of Telegram's shaders. Device-tilt motion has no
input on this desktop UI.

Format references: [Telegram themes](https://core.telegram.org/api/themes),
[wallpapers](https://core.telegram.org/api/wallpapers), and the bundled
Telegram Desktop source for behavior, palette identifiers and archive entry
names. Telegram Desktop implementation code was not copied. SVG pattern
rasterization uses the MIT-licensed `oksvg`/`rasterx` libraries.

## Language and verification

Standard chat, shared-media and gift strings map to the server's `tdesktop`
language pack. Both plain strings and plural forms are retained in the offline
cache. Named placeholders and Russian/English plural rules are handled by the
catalog; bundled translations are fallbacks. Application-specific controls
(such as local theme import paths and external player actions) keep their own catalog keys.

Tests cover independent search pagination, saved-peer scope, media references,
collection metadata/privacy, offline language packs, plural substitution,
scrollbar dragging, cache retention and tray release. For GPU screenshots:

```sh
SHARED_PNG=/tmp/shared go test ./internal/messenger/ui -run TestRenderSharedMediaAndThemes
```

The render fixture uses demo data. Authenticated Telegram account behavior
requires a manual live-account check; no account credentials are needed by the
unit or rendering tests.
