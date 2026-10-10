package user

import (
	_ "embed"

	"github.com/siti-nabila/api-contracts/pkg/dictionary"
)

var (
	//go:embed errors.yaml
	errorList []byte

	registry = dictionary.MustLoadYAML("user", errorList)

	ErrResourceBusy        = registry.MustNew("resource_busy")
	ErrIdempotencyConflict = registry.MustNew("idempotency_conflict")
)

func Registry() dictionary.Registry {
	return registry
}
