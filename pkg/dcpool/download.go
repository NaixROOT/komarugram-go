package dcpool

import (
	"context"
	"io"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

// DownloadClient keeps origin requests on the file's DC and CDN requests on
// separate, unauthenticated CDN transports managed by gotd (including RSA keys).
func DownloadClient(p Pool, ctx context.Context, dc int) downloader.Client {
	api := p.Client(ctx, dc)
	if provider, ok := p.(downloader.CDNProvider); ok {
		return downloadClient{api, provider}
	}
	return api
}

type downloadClient struct {
	*tg.Client
	downloader.CDNProvider
}

func (p *pool) CDN(ctx context.Context, dc int, max int64) (downloader.CDN, io.Closer, error) {
	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()
	if closed {
		return nil, nil, io.ErrClosedPipe
	}
	invoker, err := p.api.CDN(ctx, dc, max)
	if err != nil {
		return nil, nil, err
	}
	if p.observer != nil {
		return tg.NewClient(chainMiddlewares(invoker, p.observer(dc)...)), invoker, nil
	}
	return tg.NewClient(invoker), invoker, nil
}
