package user

import (
	_ "embed"

	"github.com/siti-nabila/api-contracts/pkg/dictionary"
)

var (
	//go:embed errors.yaml
	errorList []byte

	registry = dictionary.MustLoadYAML("user", errorList)

	ErrResourceBusy                 = registry.MustNew("resource_busy")
	ErrIdempotencyConflict          = registry.MustNew("idempotency_conflict")
	ErrVersionConflict              = registry.MustNew("version_conflict")
	ErrInvalidStatusTransition      = registry.MustNew("invalid_status_transition")
	ErrSelfAdminRoleRevokeForbidden = registry.MustNew("self_admin_role_revoke_forbidden")
	ErrSelfInactivationForbidden    = registry.MustNew("self_inactivation_forbidden")
	ErrGrantTargetInactive          = registry.MustNew("grant_target_inactive")
	ErrLastAdminRequired            = registry.MustNew("last_admin_required")
)

func Registry() dictionary.Registry {
	return registry
}
