package auth

import (
	_ "embed"

	"github.com/siti-nabila/api-contracts/pkg/dictionary"
)

var (
	//go:embed errors.yaml
	errorList []byte

	registry = dictionary.MustLoadYAML("auth", errorList)

	ErrDataExists       = registry.MustNew("already_exists")
	ErrNotFound         = registry.MustNew("not_found")
	ErrPasswordMismatch = registry.MustNew("password_mismatch")
	ErrRequired         = registry.MustNew("required")
	ErrAlphaOnly        = registry.MustNew("alpha_only")
	ErrInvalidEmail     = registry.MustNew("invalid_email")
	ErrGeneratingToken  = registry.MustNew("generate_token")
)

func Registry() dictionary.Registry {
	return registry
}

func ErrMinLength(length int) error {
	return registry.MustNewf("min_length", length)
}

func ErrMaxLength(length int) error {
	return registry.MustNewf("max_length", length)
}
