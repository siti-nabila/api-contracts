package mapping

import (
	"github.com/siti-nabila/api-contracts/pkg/dictionary/user"
	"github.com/siti-nabila/api-contracts/pkg/grpcerror"
	"google.golang.org/grpc/codes"
)

func UserCodeMappings() []grpcerror.CodeMapping {
	return []grpcerror.CodeMapping{
		{Code: codes.FailedPrecondition, Errors: []error{user.ErrInvalidStatusTransition, user.ErrLastAdminRequired}},
		{Code: codes.PermissionDenied, Errors: []error{user.ErrSelfInactivationForbidden}},
		{Code: codes.Aborted, Errors: []error{user.ErrResourceBusy, user.ErrIdempotencyConflict, user.ErrVersionConflict}},
	}
}
