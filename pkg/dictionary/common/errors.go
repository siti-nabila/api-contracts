package common

import (
	_ "embed"

	"github.com/siti-nabila/api-contracts/pkg/dictionary"
)

var (
	//go:embed errors.yaml
	errorList []byte

	registry = dictionary.MustLoadYAML("common", errorList)

	ErrBadRequest          = registry.MustNew("bad_request")
	ErrNotFound            = registry.MustNew("not_found")
	ErrUnauthorized        = registry.MustNew("unauthorized")
	ErrForbidden           = registry.MustNew("forbidden")
	ErrServiceUnavailable  = registry.MustNew("service_unavailable")
	ErrDeadlineExceeded    = registry.MustNew("deadline_exceeded")
	ErrInternalServerError = registry.MustNew("internal_server_error")
)

func Registry() dictionary.Registry {
	return registry
}
