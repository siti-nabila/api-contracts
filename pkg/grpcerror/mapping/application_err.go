package mapping

import (
	"github.com/siti-nabila/api-contracts/pkg/dictionary/application"
	"github.com/siti-nabila/api-contracts/pkg/grpcerror"
	"google.golang.org/grpc/codes"
)

func ApplicationCodeMappings() []grpcerror.CodeMapping {
	return []grpcerror.CodeMapping{
		{Code: codes.NotFound, Errors: []error{application.ErrApplicationNotFound}},
		{Code: codes.FailedPrecondition, Errors: []error{application.ErrApplicationInactive, application.ErrSystemApplicationProtected}},
	}
}
