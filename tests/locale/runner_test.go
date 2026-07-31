package locale_test

import (
	"testing"

	"github.com/siti-nabila/api-contracts/tests/locale/test_scenarios"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
)

func TestLocale(t *testing.T) {
	testutils.Run(t, test_scenarios.All())
}
