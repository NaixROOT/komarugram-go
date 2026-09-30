// SPDX-License-Identifier: Unlicense OR MIT

// Command emoji-pack makes the catalogs of emoji packs the client installs
// from: the one built into it, and a directory to publish as it is.
//
//	go run ./cmd/emoji-pack official -tdesktop COMMIT -lib-ui COMMIT -sets DIR
//
// writes internal/messenger/emojipacks/official.json, the catalog built in:
// the emoji sets of Telegram Desktop, by where Telegram keeps their files.
// The sprites of the set it is built with and the list that orders the
// cells are downloaded from its repositories at the commits given, which
// the catalog pins; the sets it downloads are read from DIR as
// set-<post>.zip, where the opt-in TestLiveEmojiSets of cmd/messenger
// leaves them. Nothing of them is kept: the catalog has their addresses,
// sizes and hashes. To move to a newer version of the sets, run it with
// the newer commits, and the newer posts in official.go.
//
// A catalog of a directory has the files of its packs:
//
//	go run ./cmd/emoji-pack -catalog DIR font -id noto -name "Noto Color Emoji" -license OFL-1.1 NotoColorEmoji.ttf
//	go run ./cmd/emoji-pack -catalog DIR telegram -id twemoji -name Twemoji -list emoji.txt DIR-OF-THE-SET
//	go run ./cmd/emoji-pack -catalog DIR telegram -id twemoji -name Twemoji -list emoji.txt -post tdhbcfiles/3224 set-3224.zip
//	go run ./cmd/emoji-pack -catalog DIR list
//	go run ./cmd/emoji-pack -catalog DIR remove noto
//
// A font pack is one emoji font. A telegram pack is an emoji set of
// Telegram Desktop: its emoji_N.webp sprites, from the set's directory,
// named one by one or in the set's zip, and lib_ui/emoji.txt of the sources
// of the version the set is of, which orders the cells. With -post the
// sprites are not put into the catalog: the zip is the file of that post of
// a Telegram channel, and the client downloads it from there through the
// user's account.
//
// The client is pointed at such a catalog, in place of the one built in, by
// KOMARUGRAM_EMOJI_PACKS, a directory or a URL.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"komarugram/internal/messenger/emojipacks"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "emoji-pack:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 0 && args[0] == "official" {
		return official(args[1:])
	}
	global := flag.NewFlagSet("emoji-pack", flag.ContinueOnError)
	catalog := global.String("catalog", "", "the catalog's directory")
	global.Usage = func() {
		fmt.Fprintln(global.Output(), "usage: emoji-pack official ... | emoji-pack -catalog DIR font|telegram|list|remove ...")
		global.PrintDefaults()
	}
	if err := global.Parse(args); err != nil {
		return err
	}
	if *catalog == "" || global.NArg() == 0 {
		global.Usage()
		return fmt.Errorf("a catalog and a command are needed")
	}
	if err := os.MkdirAll(*catalog, 0o755); err != nil {
		return err
	}
	command, rest := global.Arg(0), global.Args()[1:]
	pack := flag.NewFlagSet(command, flag.ContinueOnError)
	var p emojipacks.Pack
	pack.StringVar(&p.ID, "id", "", "the pack's id: lower-case letters, digits, '-' and '_'")
	pack.StringVar(&p.Name, "name", "", "the pack's name, shown to the user")
	pack.StringVar(&p.License, "license", "", "the pack's license, shown to the user")
	pack.StringVar(&p.Source, "source", "", "where the pack is from, shown to the user")
	list := pack.String("list", "", "telegram: lib_ui/emoji.txt of Telegram Desktop's sources")
	post := pack.String("post", "", "telegram: CHANNEL/POST whose file the zip given is; the sprites stay there")
	version := pack.Int("version", 0, "telegram: the version of the set the list is of, checked against the zip's; 0 for any")
	switch command {
	case "font":
		if err := pack.Parse(rest); err != nil {
			return err
		}
		if pack.NArg() != 1 {
			return fmt.Errorf("font: one font file is needed")
		}
		built, err := emojipacks.BuildFont(*catalog, p, pack.Arg(0))
		if err != nil {
			return err
		}
		fmt.Printf("%s: a font of %.1f MB\n", built.ID, float64(built.Size())/(1<<20))
	case "telegram":
		if err := pack.Parse(rest); err != nil {
			return err
		}
		if *list == "" || pack.NArg() == 0 {
			return fmt.Errorf("telegram: -list and the set's directory or sprites are needed")
		}
		text, err := os.ReadFile(*list)
		if err != nil {
			return err
		}
		order, err := emojipacks.TelegramOrder(string(text))
		if err != nil {
			return fmt.Errorf("%s: %w", *list, err)
		}
		var (
			images  []string
			zipPath string
			archive *emojipacks.TelegramArchive
		)
		if pack.NArg() == 1 && strings.EqualFold(filepath.Ext(pack.Arg(0)), ".zip") {
			zipPath = pack.Arg(0)
			tmp, err := os.MkdirTemp("", "emoji-pack")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmp)
			var got int
			if images, got, err = emojipacks.TelegramSprites(zipPath, tmp); err != nil {
				return err
			}
			if *version != 0 && got != *version {
				return fmt.Errorf("%s is version %d of the set, the list is of version %d", zipPath, got, *version)
			}
			fmt.Printf("%s: version %d of the set\n", filepath.Base(zipPath), got)
		} else if images, err = sprites(pack.Args()); err != nil {
			return err
		}
		if *post != "" {
			channel, number, ok := strings.Cut(*post, "/")
			n, err := strconv.Atoi(number)
			if !ok || err != nil || zipPath == "" {
				return fmt.Errorf("telegram: -post takes CHANNEL/POST, and the set as its zip")
			}
			archive = &emojipacks.TelegramArchive{Channel: channel, Post: n}
		}
		built, err := emojipacks.BuildTelegramSprites(*catalog, p, images, order, zipPath, archive)
		if err != nil {
			return err
		}
		where := "in the catalog"
		if archive != nil {
			where = "in Telegram, @" + *post
		}
		fmt.Printf("%s: %d emoji in %d sprites %s, %.1f MB to download\n", built.ID, len(order), len(images), where, float64(built.DownloadSize())/(1<<20))
	case "list":
		packs, err := emojipacks.ReadIndex(context.Background(), emojipacks.NewSource(*catalog))
		if err != nil {
			return err
		}
		for _, p := range packs {
			from := ""
			if p.Telegram != nil {
				from = fmt.Sprintf("  @%s/%d", p.Telegram.Channel, p.Telegram.Post)
			}
			fmt.Printf("%-16s %-8s %6.1f MB  %s  %s%s\n", p.ID, p.Kind, float64(p.DownloadSize())/(1<<20), p.Name, p.License, from)
		}
	case "remove":
		if len(rest) != 1 {
			return fmt.Errorf("remove: one pack id is needed")
		}
		return emojipacks.RemoveFromCatalog(*catalog, rest[0])
	default:
		global.Usage()
		return fmt.Errorf("unknown command %q", command)
	}
	return nil
}

var spriteName = regexp.MustCompile(`^emoji_(\d+)\.webp$`)

// sprites returns the sprites of a set in order: the emoji_N.webp of a
// directory, or the files named.
func sprites(args []string) ([]string, error) {
	var files []string
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			files = append(files, arg)
			continue
		}
		found, err := filepath.Glob(filepath.Join(arg, "emoji_*.webp"))
		if err != nil {
			return nil, err
		}
		files = append(files, found...)
	}
	number := func(path string) int {
		m := spriteName.FindStringSubmatch(strings.ToLower(filepath.Base(path)))
		if m == nil {
			return -1
		}
		n, _ := strconv.Atoi(m[1])
		return n
	}
	for _, f := range files {
		if number(f) < 1 {
			return nil, fmt.Errorf("%s is not a sprite of a set: emoji_N.webp", f)
		}
	}
	sort.Slice(files, func(i, j int) bool { return number(files[i]) < number(files[j]) })
	for i, f := range files {
		if number(f) != i+1 {
			return nil, fmt.Errorf("the sprites are not emoji_1.webp to emoji_%d.webp: %s", len(files), filepath.Base(f))
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no emoji_N.webp sprites")
	}
	return files, nil
}
