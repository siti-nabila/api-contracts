package application

import (
	_ "embed"
	"github.com/siti-nabila/api-contracts/pkg/dictionary"
)

var (
	//go:embed errors.yaml
	errorList                     []byte
	registry                      = dictionary.MustLoadYAML("application", errorList)
	ErrApplicationNotFound        = registry.MustNew("application_not_found")
	ErrApplicationAccessRequired  = registry.MustNew("application_access_required")
	ErrApplicationInactive        = registry.MustNew("application_inactive")
	ErrSystemApplicationProtected = registry.MustNew("system_application_protected")
)

func Registry() dictionary.Registry { return registry }
