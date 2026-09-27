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
			Name: "preserves arbitrary regional language",
			Run:  preserveArbitraryLanguage,
		},
		{
			Name: "falls back to English for malformed language",
			Run:  fallbackMalformedLanguage,
		},
		{
			Name: "stores language in request context",
			Run:  contextLanguage,
		},
	}
}

func parseRegionalLanguage(t *testing.T) {
	if got := locale.Parse("id-ID,id;q=0.9,en;q=0.8"); got != locale.Language("id-ID") {
		t.Errorf("Parse() = %q, want id-ID", got)
	}
}

func preserveArbitraryLanguage(t *testing.T) {
	if got := locale.Parse("zh-CN,zh;q=0.9"); got != locale.Language("zh-CN") {
		t.Errorf("Parse() = %q, want zh-CN", got)
	}
}

func fallbackMalformedLanguage(t *testing.T) {
	if got := locale.Parse("invalid@locale"); got != locale.DefaultLanguage {
		t.Errorf("Parse() = %q, want default %q", got, locale.DefaultLanguage)
	}
}

func contextLanguage(t *testing.T) {
	ctx := locale.NewContext(context.Background(), "id")
	if got := locale.FromContext(ctx); got != locale.Indonesian {
		t.Errorf("FromContext() = %q, want %q", got, locale.Indonesian)
	}
}
