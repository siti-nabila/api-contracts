package mapping

import (
	"github.com/siti-nabila/api-contracts/pkg/dictionary/auth"
	"github.com/siti-nabila/api-contracts/pkg/grpcerror"
	"google.golang.org/grpc/codes"
)

func AuthCodeMappings() []grpcerror.CodeMapping {
	return []grpcerror.CodeMapping{
		{
			Code: codes.InvalidArgument,
			Errors: []error{
				auth.ErrMinLength(0),
				auth.ErrMaxLength(0),
				auth.ErrAlphaOnly,
				auth.ErrInvalidEmail,
				auth.ErrRequired,
				auth.ErrPasswordMismatch,
			},
		},
		{
			Code:   codes.NotFound,
			Errors: []error{auth.ErrNotFound},
		},
		{
			Code:   codes.AlreadyExists,
			Errors: []error{auth.ErrDataExists},
		},
	}
}
