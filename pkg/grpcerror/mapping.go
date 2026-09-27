package grpcerror

import (
	"errors"

	"google.golang.org/grpc/codes"
)

type CodeMapping struct {
	Code   codes.Code
	Errors []error
}

func cloneCodeMappings(mappings []CodeMapping) []CodeMapping {
	cloned := make([]CodeMapping, len(mappings))
	for index, mapping := range mappings {
		cloned[index] = CodeMapping{
			Code:   mapping.Code,
			Errors: append([]error(nil), mapping.Errors...),
		}
	}
	return cloned
}

func resolveCode(err error, mappings []CodeMapping) (codes.Code, bool) {
	for _, mapping := range mappings {
		for _, candidate := range mapping.Errors {
			if candidate != nil && errors.Is(err, candidate) {
				return mapping.Code, true
			}
		}
	}
	return codes.Internal, false
}
