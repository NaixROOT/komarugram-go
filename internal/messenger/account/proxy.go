// SPDX-License-Identifier: Unlicense OR MIT

package account

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/gotd/td/telegram/dcs"
)

// ParseProxy makes a resolver that connects through the MTProxy in a
// tg://proxy?server=…&port=…&secret=… link, or the same link on t.me.
func ParseProxy(link string) (dcs.Resolver, error) {
	u, err := url.Parse(strings.TrimSpace(link))
	if err != nil {
		return nil, fmt.Errorf("account: proxy link: %w", err)
	}
	isProxy := (u.Scheme == "tg" && u.Host == "proxy") ||
		((u.Scheme == "https" || u.Scheme == "http") && (u.Host == "t.me" || u.Host == "telegram.me") && u.Path == "/proxy")
	if !isProxy {
		return nil, errors.New("account: proxy link must look like tg://proxy?server=…&port=…&secret=…")
	}
	q := u.Query()
	server, port, secret := q.Get("server"), q.Get("port"), q.Get("secret")
	if server == "" || port == "" || secret == "" {
		return nil, errors.New("account: proxy link needs server, port and secret")
	}
	key, err := decodeSecret(secret)
	if err != nil {
		return nil, err
	}
	return dcs.MTProxy(net.JoinHostPort(server, port), key, dcs.MTProxyOptions{})
}

// decodeSecret reads a proxy secret written in hex or in base64.
func decodeSecret(s string) ([]byte, error) {
	if key, err := hex.DecodeString(s); err == nil {
		return key, nil
	}
	for _, enc := range []*base64.Encoding{base64.RawURLEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.StdEncoding} {
		if key, err := enc.DecodeString(s); err == nil {
			return key, nil
		}
	}
	return nil, errors.New("account: proxy secret is neither hex nor base64")
}
