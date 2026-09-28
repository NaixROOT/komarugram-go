package localization

import (
	"strings"
	"testing"
)

func TestCatalogSelectionAndFallback(t *testing.T) {
	if got := For("en").T("nav.settings"); got != "Settings" {
		t.Fatalf("English settings = %q", got)
	}
	if got := For("unknown").T("nav.settings"); got != "Настройки" {
		t.Fatalf("fallback settings = %q", got)
	}
	if got := For("en").T("missing.key"); got != "missing.key" {
		t.Fatalf("missing key = %q", got)
	}
}

// A key with an English text and no Russian one shows as itself in the
// Russian interface.
func TestEveryEnglishKeyHasRussian(t *testing.T) {
	for key := range english {
		if strings.Contains(key, "#") {
			continue // Plural forms differ between the languages.
		}
		if russian[key] == "" {
			t.Errorf("%s has no Russian text", key)
		}
	}
}
