// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"
	"io"
	"komarugram/pkg/dcpool"
	"net/http"
	"net/url"
	"time"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

func downloadPickerWeb(ctx context.Context, pool dcpool.Pool, ref fileLocation) ([]byte, error) {
	out := &cappedBuffer{}
	if !ref.WebNoProxy {
		_, err := downloader.NewDownloader().Web(dcpool.DownloadClient(pool, ctx, ref.DC), &tg.InputWebFileLocation{URL: ref.WebURL, AccessHash: ref.WebHash}).Stream(ctx, out)
		return out.Bytes(), err
	}
	check := func(raw string) error {
		u, e := url.Parse(raw)
		if e != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
			return errors.New("invalid GIF media URL")
		}
		return nil
	}
	if err := check(ref.WebURL); err != nil {
		return nil, err
	}
	client := http.Client{Timeout: 45 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) > 5 {
			return errors.New("too many GIF redirects")
		}
		return check(req.URL.String())
	}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ref.WebURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("GIF download failed: " + res.Status)
	}
	if res.ContentLength > MaxMediaBytes {
		return nil, errors.New("GIF exceeds media limit")
	}
	_, err = io.Copy(out, io.LimitReader(res.Body, MaxMediaBytes+1))
	return out.Bytes(), err
}
