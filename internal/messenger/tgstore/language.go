// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"komarugram/internal/crash"
	"time"

	"github.com/gotd/td/tg"
)

// LanguagePack overlays server strings, including plural forms, with an
// account-scoped offline copy. Failures have a bounded retry interval.
func (s *Store) LanguagePack(language string) map[string]string {
	c := s.history
	c.mu.Lock()
	defer c.mu.Unlock()
	pack := c.packs[language]
	if c.closing || c.cache == nil || c.packLoading[language] || time.Now().Before(c.packNext[language]) {
		return pack
	}
	if c.api == nil && pack != nil {
		return pack
	}
	c.packLoading[language] = true
	api, cache, parent := c.api, c.cache, c.ctx
	c.wg.Go(func() {
		retry := time.Minute
		defer func() {
			c.mu.Lock()
			c.packLoading[language] = false
			if c.packNext == nil {
				c.packNext = map[string]time.Time{}
			}
			c.packNext[language] = time.Now().Add(retry)
			c.mu.Unlock()
		}()
		defer crash.Recover("language pack", func(*crash.Panic) {})
		ctx, cancel := context.WithTimeout(parent, 15*time.Second)
		defer cancel()
		cached := map[string]string{}
		if _, err := cache.Get(ctx, "language/"+language, &cached); err == nil {
			c.mu.Lock()
			c.packs[language] = cached
			c.mu.Unlock()
			s.changed()
		}
		if api == nil {
			return
		}
		result, err := api.LangpackGetLangPack(ctx, &tg.LangpackGetLangPackRequest{LangPack: "tdesktop", LangCode: language})
		if err != nil {
			return
		}
		values := languageStrings(result.Strings)
		if err := cache.Put(ctx, "language/"+language, values); err != nil {
			return
		}
		c.mu.Lock()
		c.packs[language] = values
		c.mu.Unlock()
		retry = time.Hour
		s.changed()
	})
	return pack
}
func languageStrings(strings []tg.LangPackStringClass) map[string]string {
	values := map[string]string{}
	for _, raw := range strings {
		switch s := raw.(type) {
		case *tg.LangPackString:
			values[s.Key] = s.Value
		case *tg.LangPackStringPluralized:
			for suffix, value := range map[string]string{"zero": s.ZeroValue, "one": s.OneValue, "two": s.TwoValue, "few": s.FewValue, "many": s.ManyValue, "other": s.OtherValue} {
				if value != "" {
					values[s.Key+"#"+suffix] = value
				}
			}
		}
	}
	return values
}
