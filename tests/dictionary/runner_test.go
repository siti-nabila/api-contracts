package dictionary_test

import (
	"testing"

	"github.com/siti-nabila/api-contracts/tests/dictionary/test_scenarios"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
)

func TestDictionary(t *testing.T) {
	testutils.Run(t, test_scenarios.All())
}
