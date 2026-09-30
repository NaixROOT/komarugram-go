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

A chat is drawn in the first of: the theme set for it only here, its own
Telegram theme and wallpaper, and the look of every chat from the settings.

**Every chat** (Settings → Chat Settings, `settings_chats.go`), as Telegram
Desktop's Chat Settings: the application's own Material colors, or Classic,
Day, Tinted and Night, with an accent of the theme's eight; and a wallpaper
from Telegram's gallery (`account.getWallPapers`, cached offline) or a
file, previewed over a chat, a picture with "Blurred", before it is
applied. The light and the dark application theme each keep their own
theme and wallpaper (`preferences.ChatLook`), as Telegram Desktop keeps a
day and a night theme; choosing a dark theme turns the application dark.
The presets' chat colors are the ones of Telegram Desktop's embedded
palettes (`chattheme/presets.go`); an accent turns the colors near the
theme's own accent to its hue, grays kept, as Telegram Desktop colors a
theme. Classic's wallpaper is its four colors over a pattern of doodles
drawn for this client (`chattheme/pattern.go`): Telegram Desktop's bundled
pattern is not used. A wallpaper's picture is kept once, by its hash, in
`wallpapers` beside the settings, and removed when no look uses it; it is
not encrypted, as the settings are not.

**One chat** (the chat's information → Change colors, also in the menu of
the chat's header), as Telegram Desktop's theme chooser: the chat as it
will look, the themes as cards of their wallpaper and bubbles, and "No
theme". A card previews the theme on the chat at once; it is applied for
all in the chat (`messages.setChatTheme`, Telegram's themes and none) or
only here (the account cache), which "Remove my theme" undoes. A Telegram
Desktop theme (`.tdesktop-theme` archive or palette) imports its chat
colors and wallpaper, only here.

Telegram's chat themes are drawn as Telegram Desktop draws cloud themes:
the preset of their base theme turned to their accent, with their message
colors. A chat's appearance comes from the cache at once and is asked for
again in the background once a session (`tgstore.ChatAppearance`); a
service message that changed it is waited for.

The wallpaper is decoded once (`chattheme.Prepare`) and rendered off the
frame at the size the history takes (`Background.Render`, up to 2048 px a
side): colors, a freeform gradient of up to four colors, a picture cut to
cover or tiled, or a pattern (PNG or TGV) over the colors, darkening them
as a soft light, or, for dark wallpapers, showing them through the pattern
only. SVG patterns are rasterized within each path's bounds
(`chattheme/scanner.go`); rasterx's scanner covers the whole image for
every path. Dates, service messages and the times of stickers lie on
plates of the wallpaper's average hue. A hidden window drops its wallpaper.
The freeform gradient is implemented independently after Telegram's
published behavior; it is not a pixel-exact reproduction of its shaders.
Device-tilt motion has no input on this desktop UI.

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
