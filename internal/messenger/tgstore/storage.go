// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"

	"komarugram/internal/messenger/model"
)

var _ model.StorageUsageSource = (*Store)(nil)

// StorageUsage implements model.StorageUsageSource from the cache alone.
func (s *Store) StorageUsage(ctx context.Context) (model.CacheUsage, error) {
	cache := s.Cache()
	if cache == nil {
		return model.CacheUsage{}, errors.New("storage: cache not open")
	}
	return cache.StorageUsage(ctx)
}
