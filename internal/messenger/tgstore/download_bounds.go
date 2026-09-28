// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"io"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

// Parallel downloaders probe past EOF to discover a file's length. A CDN can
// respond to such a probe with ReuploadNeeded, then VOLUME_LOC_NOT_FOUND. Use
// Telegram's known byte length to stop those probes before they reach the RPC.
type sizedDownload struct {
	downloader.Client
	size int64
}

func (s sizedDownload) UploadGetFile(ctx context.Context, r *tg.UploadGetFileRequest) (tg.UploadFileClass, error) {
	if s.size > 0 && r.Offset >= s.size {
		return &tg.UploadFile{Type: &tg.StorageFileUnknown{}}, nil
	}
	return s.Client.UploadGetFile(ctx, r)
}

type sizedCDNDownload struct {
	sizedDownload
	provider downloader.CDNProvider
}

func (s sizedCDNDownload) CDN(ctx context.Context, dc int, max int64) (downloader.CDN, io.Closer, error) {
	client, close, err := s.provider.CDN(ctx, dc, max)
	if err != nil {
		return nil, nil, err
	}
	return sizedCDN{client, s.size}, close, nil
}

type sizedCDN struct {
	downloader.CDN
	size int64
}

func (s sizedCDN) UploadGetCDNFile(ctx context.Context, r *tg.UploadGetCDNFileRequest) (tg.UploadCDNFileClass, error) {
	if s.size > 0 && r.Offset >= s.size {
		return &tg.UploadCDNFile{}, nil
	}
	return s.CDN.UploadGetCDNFile(ctx, r)
}
func withDownloadSize(client downloader.Client, size int64) downloader.Client {
	base := sizedDownload{client, size}
	if provider, ok := client.(downloader.CDNProvider); ok {
		return sizedCDNDownload{base, provider}
	}
	return base
}
