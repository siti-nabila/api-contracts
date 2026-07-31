package grpcerror_test

import (
	"testing"

	"github.com/siti-nabila/api-contracts/tests/grpcerror/test_scenarios"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
)

func TestGRPCError(t *testing.T) {
	testutils.Run(t, test_scenarios.All())
}
