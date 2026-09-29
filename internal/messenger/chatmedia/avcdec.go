// SPDX-License-Identifier: Unlicense OR MIT

package chatmedia

import (
	"context"

	"komarugram/internal/messenger/wasmmodule"
	"komarugram/pkg/h264"
)

// newAVCRuntime compiles the H.264 decoder, fetching it first if needed.
func newAVCRuntime(ctx context.Context) (*h264.Runtime, error) {
	module, err := wasmmodule.AVCDec.Load(ctx)
	if err != nil {
		return nil, err
	}
	return h264.NewRuntime(ctx, module)
}
