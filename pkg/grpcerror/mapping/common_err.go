package mapping

import (
	"github.com/siti-nabila/api-contracts/pkg/dictionary/common"
	"github.com/siti-nabila/api-contracts/pkg/grpcerror"
	"google.golang.org/grpc/codes"
)

func CommonCodeMappings() []grpcerror.CodeMapping {
	return []grpcerror.CodeMapping{
		{
			Code: codes.InvalidArgument,
			Errors: []error{
				common.ErrBadRequest,
			},
		},
		{
			Code:   codes.NotFound,
			Errors: []error{common.ErrNotFound},
		},
		{
			Code:   codes.Unimplemented,
			Errors: []error{common.ErrEndpointNotFound, common.ErrServiceUnavailable},
		},
		{
			Code:   codes.Unauthenticated,
			Errors: []error{common.ErrUnauthorized},
		},
		{
			Code:   codes.PermissionDenied,
			Errors: []error{common.ErrForbidden},
		},
		{
			Code:   codes.DeadlineExceeded,
			Errors: []error{common.ErrDeadlineExceeded},
		},
		{
			Code:   codes.Internal,
			Errors: []error{common.ErrInternalServerError},
		},
	}
}
