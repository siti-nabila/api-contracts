package grpcerror

import (
	"github.com/siti-nabila/api-contracts/pkg/dictionary"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Decoded struct {
	Code        codes.Code
	Message     string
	Reason      string
	Metadata    map[string]string
	FieldErrors map[string][]string
}

func Decode(err error) (Decoded, bool) {
	grpcStatus, ok := status.FromError(err)
	if !ok {
		return Decoded{}, false
	}

	decoded := Decoded{
		Code:        grpcStatus.Code(),
		Message:     grpcStatus.Message(),
		Metadata:    make(map[string]string),
		FieldErrors: make(map[string][]string),
	}
	for _, detail := range grpcStatus.Details() {
		switch typed := detail.(type) {
		case *errdetails.ErrorInfo:
			decoded.Reason = typed.GetReason()
			for key, value := range typed.GetMetadata() {
				decoded.Metadata[key] = value
			}
		case *errdetails.BadRequest:
			for _, violation := range typed.GetFieldViolations() {
				field := violation.GetField()
				decoded.FieldErrors[field] = append(
					decoded.FieldErrors[field],
					violation.GetDescription(),
				)
			}
		}
	}
	return decoded, true
}

func (decoded Decoded) HasFieldErrors() bool {
	return len(decoded.FieldErrors) > 0
}

func (decoded Decoded) Definition(
	registries ...dictionary.Registry,
) (dictionary.Definition, bool) {
	for _, registry := range registries {
		if definition, exists := registry.Lookup(decoded.Reason); exists {
			return definition, true
		}
	}
	return dictionary.Definition{}, false
}
