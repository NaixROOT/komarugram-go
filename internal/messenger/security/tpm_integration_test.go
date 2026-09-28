// SPDX-License-Identifier: Unlicense OR MIT

package security

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// This test is opt-in because it consumes real TPM commands and should never
// submit an intentionally wrong authorization (which would advance the TPM's
// dictionary-attack counter).
func TestHardwareTPMRoundTrip(t *testing.T) {
	if os.Getenv("KOMARUGRAM_TEST_TPM") != "1" {
		t.Skip("set KOMARUGRAM_TEST_TPM=1 to use the machine TPM")
	}
	path := filepath.Join(t.TempDir(), "security.json")
	m, err := OpenPath(path, hardwareTPM{})
	if err != nil {
		t.Fatal(err)
	}
	if !m.State().Available {
		t.Fatal("TPM probe failed")
	}
	const password = "komarugram-integration-only"
	if err := m.Enable(context.Background(), password); err != nil {
		t.Fatal(err)
	}
	restarted, err := OpenPath(path, hardwareTPM{})
	if err != nil {
		t.Fatal(err)
	}
	if err := restarted.Unlock(context.Background(), password); err != nil {
		t.Fatal(err)
	}
	key, err := restarted.DeriveKey("integration")
	if err != nil || len(key) != keySize {
		t.Fatalf("derived key length %d, err %v", len(key), err)
	}
	clear(key)
}
