package protobuf_test

import (
	"github.com/siti-nabila/api-contracts/tests/protobuf/test_scenarios"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"testing"
)

func TestSessionContract(t *testing.T) {
	testutils.Run(t, test_scenarios.All())
}
