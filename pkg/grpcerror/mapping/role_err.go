package mapping

import (
	"github.com/siti-nabila/api-contracts/pkg/dictionary/role"
	"github.com/siti-nabila/api-contracts/pkg/grpcerror"
	"google.golang.org/grpc/codes"
)

func RoleCodeMappings() []grpcerror.CodeMapping {
	return []grpcerror.CodeMapping{
		{Code: codes.NotFound, Errors: []error{role.ErrRoleNotFound}},
		{Code: codes.AlreadyExists, Errors: []error{role.ErrRoleCodeAlreadyExists}},
		{Code: codes.FailedPrecondition, Errors: []error{role.ErrSystemRoleProtected, role.ErrRoleDisabled}},
		{Code: codes.InvalidArgument, Errors: []error{role.ErrInvalidPermissionSet}},
		{Code: codes.Aborted, Errors: []error{role.ErrVersionConflict, role.ErrIdempotencyConflict, role.ErrResourceBusy}},
	}
}
