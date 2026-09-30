// SPDX-License-Identifier: Unlicense OR MIT

package audiotag

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
	"unicode/utf16"
)

func write(t *testing.T, name string, parts ...[]byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, bytes.Join(parts, nil), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// mp3Frames are n silent frames of MPEG-1 layer III, 128 kbit/s at 44.1
// kHz: 1152 samples each.
func mp3Frames(n int) []byte {
	frame := make([]byte, 417)
	copy(frame, []byte{0xff, 0xfb, 0x90, 0x64})
	return bytes.Repeat(frame, n)
}

func mp3Length(n int) time.Duration {
	return time.Duration(n*1152) * time.Second / 44100
}

func synchsafeBytes(n int) []byte {
	return []byte{byte(n >> 21 & 0x7f), byte(n >> 14 & 0x7f), byte(n >> 7 & 0x7f), byte(n & 0x7f)}
}

// id3v2 is a tag of version with frames.
func id3v2(version byte, flags byte, frames ...[]byte) []byte {
	body := bytes.Join(frames, nil)
	return append(append([]byte{'I', 'D', '3', version, 0, flags}, synchsafeBytes(len(body))...), body...)
}

func frame3(id string, data []byte) []byte {
	return append(append([]byte(id), binary.BigEndian.AppendUint32(nil, uint32(len(data)))...), append([]byte{0, 0}, data...)...)
}

func frame4(id string, flags byte, data []byte) []byte {
	return append(append([]byte(id), synchsafeBytes(len(data))...), append([]byte{0, flags}, data...)...)
}

func frame2(id string, data []byte) []byte {
	n := len(data)
	return append(append([]byte(id), byte(n>>16), byte(n>>8), byte(n)), data...)
}

func utf16LE(s string) []byte {
	b := []byte{0xff, 0xfe}
	for _, u := range utf16.Encode([]rune(s)) {
		b = binary.LittleEndian.AppendUint16(b, u)
	}
	return b
}

var (
	frontCover = []byte("\x89PNG front cover")
	backCover  = []byte("\xff\xd8 back cover")
)

func TestMP3WithID3v23(t *testing.T) {
	apic := func(kind byte, data []byte) []byte {
		// UTF-16 description, ended by two zeroes.
		return append(append(append([]byte{1}, "image/png\x00"...), kind), append(append(utf16LE("обложка"), 0, 0), data...)...)
	}
	path := write(t, "a.mp3", id3v2(3, 0,
		frame3("TIT2", append([]byte{1}, utf16LE("Ангел")...)),
		frame3("TPE1", []byte("\x00Caf\xe9")),
		frame3("APIC", apic(4, backCover)),
		frame3("APIC", apic(3, frontCover)),
		frame3("APIC", apic(0, backCover)),
		make([]byte, 20), // padding
	), mp3Frames(40))
	info, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Title != "Ангел" || info.Performer != "Café" || !bytes.Equal(info.Cover, frontCover) || info.Duration != mp3Length(40) {
		t.Fatalf("%+v", info)
	}
}

// ID3v2.4: UTF-8, several values, a frame unsynchronised with its data
// length before it.
func TestMP3WithID3v24(t *testing.T) {
	cover := []byte{0xff, 0xd8, 0xff, 0xe0, 1, 2}
	unsynced := []byte{0xff, 0x00, 0xd8, 0xff, 0x00, 0xe0, 1, 2}
	apic := append(append([]byte{0}, "image/jpeg\x00\x03\x00"...), unsynced...)
	path := write(t, "b.mp3", id3v2(4, 0,
		frame4("TIT2", 0, []byte("\x03Song\x00")),
		frame4("TPE1", 0, []byte("\x03Ann\x00Bob")),
		frame4("APIC", 0x03, append(synchsafeBytes(len(cover)+14), apic...)),
	), mp3Frames(3))
	info, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Title != "Song" || info.Performer != "Ann, Bob" || !bytes.Equal(info.Cover, cover) {
		t.Fatalf("%+v", info)
	}
}

func TestMP3WithID3v22AndV1(t *testing.T) {
	path := write(t, "c.mp3", id3v2(2, 0,
		frame2("TT2", []byte("\x00Old")),
		frame2("PIC", append([]byte("\x00PNG\x03desc\x00"), frontCover...)),
	), mp3Frames(5))
	info, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Title != "Old" || info.Performer != "" || !bytes.Equal(info.Cover, frontCover) {
		t.Fatalf("%+v", info)
	}
	v1 := make([]byte, 128)
	copy(v1, "TAG")
	copy(v1[3:], "Only v1")
	copy(v1[33:], "Somebody")
	info, err = Read(write(t, "d.mp3", mp3Frames(5), v1))
	if err != nil {
		t.Fatal(err)
	}
	if info.Title != "Only v1" || info.Performer != "Somebody" || info.Duration != mp3Length(5) {
		t.Fatalf("%+v", info)
	}
}

func flacBlock(kind byte, last bool, data []byte) []byte {
	if last {
		kind |= 0x80
	}
	n := len(data)
	return append([]byte{kind, byte(n >> 16), byte(n >> 8), byte(n)}, data...)
}

func streamInfo(rate, samples uint64) []byte {
	b := make([]byte, 34)
	binary.BigEndian.PutUint64(b[10:], rate<<44|1<<41|15<<36|samples)
	return b
}

func vorbisComments(values ...string) []byte {
	b := binary.LittleEndian.AppendUint32(nil, 6)
	b = append(b, "vendor"...)
	b = binary.LittleEndian.AppendUint32(b, uint32(len(values)))
	for _, v := range values {
		b = binary.LittleEndian.AppendUint32(b, uint32(len(v)))
		b = append(b, v...)
	}
	return b
}

func pictureBlock(kind uint32, data []byte) []byte {
	b := binary.BigEndian.AppendUint32(nil, kind)
	b = binary.BigEndian.AppendUint32(b, 9)
	b = append(b, "image/png"...)
	b = binary.BigEndian.AppendUint32(b, 0)
	b = append(b, make([]byte, 16)...)
	b = binary.BigEndian.AppendUint32(b, uint32(len(data)))
	return append(b, data...)
}

func TestFLAC(t *testing.T) {
	path := write(t, "e.flac", []byte("fLaC"),
		flacBlock(0, false, streamInfo(44100, 44100*75)),
		flacBlock(6, false, pictureBlock(0, backCover)),
		flacBlock(4, false, vorbisComments("title=Night", "ARTIST=Duo A", "Artist=Duo B", "GENRE=x")),
		flacBlock(6, false, pictureBlock(3, frontCover)),
		flacBlock(1, true, make([]byte, 100)),
	)
	info, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Title != "Night" || info.Performer != "Duo A, Duo B" || !bytes.Equal(info.Cover, frontCover) || info.Duration != 75*time.Second {
		t.Fatalf("%+v", info)
	}
}

// oggPage is a page of stream 7 holding segments.
func oggPage(granule int64, segments []byte, body []byte) []byte {
	h := append([]byte("OggS\x00\x00"), binary.LittleEndian.AppendUint64(nil, uint64(granule))...)
	h = binary.LittleEndian.AppendUint32(h, 7)
	h = append(h, make([]byte, 8)...) // sequence, checksum
	h = append(h, byte(len(segments)))
	return append(append(h, segments...), body...)
}

// lacing is the segment table of one packet of n bytes.
func lacing(n int) []byte {
	s := bytes.Repeat([]byte{255}, n/255)
	return append(s, byte(n%255))
}

func TestOggOpus(t *testing.T) {
	head := append([]byte("OpusHead\x01\x02"), binary.LittleEndian.AppendUint16(nil, 312)...)
	head = append(head, make([]byte, 7)...)
	picture := base64.StdEncoding.EncodeToString(pictureBlock(3, frontCover))
	tags := append([]byte("OpusTags"), vorbisComments("TITLE=Opus song", "ARTIST=Voice", "METADATA_BLOCK_PICTURE="+picture, string(make([]byte, 600)))...)
	// The tags go over two pages: the first ends in the middle of them.
	segs := lacing(len(tags))
	path := write(t, "f.opus",
		oggPage(0, lacing(len(head)), head),
		oggPage(0, segs[:2], tags[:510]),
		oggPage(-1, segs[2:], tags[510:]),
		oggPage(48000*10+312, []byte{3}, []byte{1, 2, 3}),
	)
	info, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Title != "Opus song" || info.Performer != "Voice" || !bytes.Equal(info.Cover, frontCover) || info.Duration != 10*time.Second {
		t.Fatalf("%+v", info)
	}
}

func TestOggVorbis(t *testing.T) {
	id := append([]byte("\x01vorbis"), make([]byte, 23)...)
	binary.LittleEndian.PutUint32(id[12:], 22050)
	tags := append(append([]byte("\x03vorbis"), vorbisComments("TITLE=Vorbis")...), 1)
	path := write(t, "g.ogg",
		oggPage(0, lacing(len(id)), id),
		oggPage(0, lacing(len(tags)), tags),
		oggPage(22050*3, []byte{1}, []byte{0}),
		oggPage(-1, []byte{1}, []byte{0}),
	)
	info, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Title != "Vorbis" || info.Duration != 3*time.Second {
		t.Fatalf("%+v", info)
	}
}

func box(kind string, parts ...[]byte) []byte {
	body := bytes.Join(parts, nil)
	return append(append(binary.BigEndian.AppendUint32(nil, uint32(8+len(body))), kind...), body...)
}

func dataBox(kind uint32, value []byte) []byte {
	return box("data", binary.BigEndian.AppendUint32(nil, kind), make([]byte, 4), value)
}

func TestM4A(t *testing.T) {
	mvhd := make([]byte, 100)
	binary.BigEndian.PutUint32(mvhd[12:], 1000)
	binary.BigEndian.PutUint32(mvhd[16:], 61500)
	path := write(t, "h.m4a",
		box("ftyp", []byte("M4A \x00\x00\x00\x00")),
		box("mdat", make([]byte, 64)),
		box("moov", box("mvhd", mvhd), box("udta", box("meta", make([]byte, 4), box("hdlr", make([]byte, 25)), box("ilst",
			box("\xa9nam", dataBox(1, []byte("Мелодия"))),
			box("\xa9ART", dataBox(1, []byte("Оркестр"))),
			box("covr", dataBox(14, frontCover), dataBox(13, backCover)),
		)))),
	)
	info, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Title != "Мелодия" || info.Performer != "Оркестр" || !bytes.Equal(info.Cover, frontCover) || info.Duration != 61500*time.Millisecond {
		t.Fatalf("%+v", info)
	}
}

func TestAAC(t *testing.T) {
	// ADTS frames at 44.1 kHz, of one raw block each.
	frame := make([]byte, 200)
	copy(frame, []byte{0xff, 0xf1, 0x50, 0x80, byte(200 >> 3), byte(200&7)<<5 | 0x1f, 0xfc})
	path := write(t, "i.aac", id3v2(3, 0, frame3("TIT2", []byte("\x00Raw AAC"))), bytes.Repeat(frame, 441))
	info, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Title != "Raw AAC" || info.Duration != time.Duration(441*1024)*time.Second/44100 {
		t.Fatalf("%+v", info)
	}
}

func TestWhatIsNoMusic(t *testing.T) {
	for _, c := range []struct{ name, data string }{
		{"a.mp3", "not an mp3 at all"},
		{"b.flac", "fLaC"},
		{"c.ogg", "OggS but nothing more"},
		{"d.m4a", "no boxes"},
		{"e.aac", "no frames"},
	} {
		if _, err := Read(write(t, c.name, []byte(c.data))); err == nil {
			t.Errorf("%s was read as music", c.name)
		}
	}
	if _, err := Read(write(t, "f.wav", []byte("RIFF"))); !errors.Is(err, ErrFormat) {
		t.Errorf("wav: %v", err)
	}
}
