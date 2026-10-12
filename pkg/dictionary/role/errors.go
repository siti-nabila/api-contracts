package role

import (
	_ "embed"
	"github.com/siti-nabila/api-contracts/pkg/dictionary"
)

var (
	//go:embed errors.yaml
	errorList                []byte
	registry                 = dictionary.MustLoadYAML("role", errorList)
	ErrRoleNotFound          = registry.MustNew("role_not_found")
	ErrRoleCodeAlreadyExists = registry.MustNew("role_code_already_exists")
	ErrSystemRoleProtected   = registry.MustNew("system_role_protected")
	ErrInvalidPermissionSet  = registry.MustNew("invalid_permission_set")
	ErrVersionConflict       = registry.MustNew("version_conflict")
	ErrIdempotencyConflict   = registry.MustNew("idempotency_conflict")
	ErrResourceBusy          = registry.MustNew("resource_busy")
)

func Registry() dictionary.Registry { return registry }
