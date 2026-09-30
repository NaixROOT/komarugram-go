// SPDX-License-Identifier: Unlicense OR MIT

package model

// ContentReader is a Store that marks voice messages listened to.
type ContentReader interface {
	// ReadContents tells that the account listened to voice message m,
	// which came with MediaUnread: the message loses the mark, and its
	// sender learns of it as far as Ghost.SendRead allows.
	ReadContents(m Message)
}

// WaveformKeeper is a Store that keeps the waveform worked out for a voice
// message that came without one.
type WaveformKeeper interface {
	KeepWaveform(m Message, waveform []byte)
}
