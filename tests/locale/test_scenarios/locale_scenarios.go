package test_scenarios

import (
	"context"
	"testing"

	"github.com/siti-nabila/api-contracts/pkg/locale"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
)

func All() []testutils.Scenario {
	return []testutils.Scenario{
		{
			Name: "parses regional Accept-Language value",
			Run:  parseRegionalLanguage,
		},
		{
			Name: "falls back to English for unsupported language",
			Run:  fallbackLanguage,
		},
		{
			Name: "stores language in request context",
			Run:  contextLanguage,
		},
	}
}

func parseRegionalLanguage(t *testing.T) {
	if got := locale.Parse("id-ID,id;q=0.9,en;q=0.8"); got != locale.Indonesian {
		t.Errorf("Parse() = %q, want %q", got, locale.Indonesian)
	}
}

func fallbackLanguage(t *testing.T) {
	if got := locale.Parse("fr-FR"); got != locale.DefaultLanguage {
		t.Errorf("Parse() = %q, want default %q", got, locale.DefaultLanguage)
	}
}

func contextLanguage(t *testing.T) {
	ctx := locale.NewContext(context.Background(), "id")
	if got := locale.FromContext(ctx); got != locale.Indonesian {
		t.Errorf("FromContext() = %q, want %q", got, locale.Indonesian)
	}
}
