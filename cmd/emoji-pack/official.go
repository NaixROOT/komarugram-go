// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"komarugram/internal/messenger/emojipacks"
)

// officialSets are the emoji sets of Telegram Desktop, as its
// chat_helpers/emoji_sets_manager.cpp lists them: the one it is built with,
// whose sprites are in its repository, and those it downloads, each the
// file of a post of a channel.
var officialSets = []struct {
	pack emojipacks.Pack
	post int
}{
	{emojipacks.Pack{ID: "apple", Name: "Apple", License: "Apple"}, 0},
	{emojipacks.Pack{ID: "android", Name: "Android", License: "Apache-2.0"}, 3223},
	{emojipacks.Pack{ID: "twemoji", Name: "Twemoji", License: "CC-BY-4.0"}, 3224},
	{emojipacks.Pack{ID: "joypixels", Name: "JoyPixels", License: "JoyPixels Free License"}, 3225},
}

const (
	officialChannel = "tdhbcfiles"
	// The sprites of the set Telegram Desktop is built with, and how many.
	officialSprites     = "https://raw.githubusercontent.com/telegramdesktop/tdesktop/%s/Telegram/Resources/emoji/emoji_%d.webp"
	officialSpriteCount = 8
	officialList        = "https://raw.githubusercontent.com/desktop-app/lib_ui/%s/emoji.txt"
)

// official writes the catalog built into the client.
func official(args []string) error {
	flags := flag.NewFlagSet("official", flag.ContinueOnError)
	tdesktop := flags.String("tdesktop", "", "the commit of telegramdesktop/tdesktop to take the sprites of its set at")
	libUI := flags.String("lib-ui", "", "the commit of desktop-app/lib_ui, its Telegram/lib_ui at that commit, to take emoji.txt at")
	sets := flags.String("sets", "", "the directory of the sets it downloads, set-<post>.zip: what TestLiveEmojiSets of cmd/messenger leaves")
	out := flags.String("out", filepath.FromSlash("internal/messenger/emojipacks/official.json"), "the file to write")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if len(*tdesktop) != 40 || len(*libUI) != 40 || *sets == "" {
		return fmt.Errorf("official: -tdesktop and -lib-ui, full commit hashes, and -sets are needed")
	}
	list := emojipacks.OfficialFile{Name: "emoji.txt", URL: fmt.Sprintf(officialList, *libUI)}
	var err error
	if list.Data, err = download(list.URL); err != nil {
		return err
	}
	var packs []emojipacks.Pack
	for _, set := range officialSets {
		p := set.pack
		p.Source = "Telegram Desktop"
		if set.post == 0 {
			var images []emojipacks.OfficialFile
			for n := 1; n <= officialSpriteCount; n++ {
				f := emojipacks.OfficialFile{Name: fmt.Sprintf("emoji_%d.webp", n), URL: fmt.Sprintf(officialSprites, *tdesktop, n)}
				if f.Data, err = download(f.URL); err != nil {
					return err
				}
				images = append(images, f)
			}
			if p, err = emojipacks.OfficialPack(p, list, images, nil, nil); err != nil {
				return err
			}
		} else {
			archive, err := os.ReadFile(filepath.Join(*sets, fmt.Sprintf("set-%d.zip", set.post)))
			if err != nil {
				return err
			}
			if p, err = emojipacks.OfficialPack(p, list, nil, archive, &emojipacks.TelegramArchive{Channel: officialChannel, Post: set.post}); err != nil {
				return err
			}
		}
		fmt.Printf("%s: %d sprites, %.1f MB to download\n", p.ID, len(p.Sprites.Images), float64(p.DownloadSize())/(1<<20))
		packs = append(packs, p)
	}
	return emojipacks.WriteIndex(*out, packs)
}

func download(address string) ([]byte, error) {
	resp, err := http.Get(address)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", address, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, emojipacks.MaxArchiveSize))
}
