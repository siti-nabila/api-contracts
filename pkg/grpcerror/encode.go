package grpcerror

import (
	"errors"
	"sort"
	"strconv"

	"github.com/siti-nabila/api-contracts/pkg/dictionary"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/common"
	"github.com/siti-nabila/api-contracts/pkg/locale"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const errorDomain = "errors.siti-nabila.dev"

func Encode(err error, language locale.Language) error {
	if err == nil {
		return nil
	}

	if grpcStatus, ok := status.FromError(err); ok {
		if len(grpcStatus.Details()) > 0 {
			return err
		}
		return encodeGRPCStatus(grpcStatus.Code(), language)
	}

	if fieldErrors, ok := errors.AsType[dictionary.FieldErrors](err); ok {
		return encodeFieldErrors(fieldErrors, language)
	}

	if applicationError, ok := errors.AsType[*dictionary.Error](err); ok {
		return encodeApplicationError(applicationError, language)
	}

	return status.Error(
		codes.Internal,
		common.ErrInternalServerError.Message(language),
	)
}

func encodeGRPCStatus(code codes.Code, language locale.Language) error {
	switch code {
	case codes.InvalidArgument:
		return encodeApplicationError(common.ErrBadRequest, language)
	case codes.Unauthenticated:
		return encodeApplicationError(common.ErrUnauthorized, language)
	case codes.PermissionDenied:
		return encodeApplicationError(common.ErrForbidden, language)
	case codes.NotFound:
		return encodeApplicationError(common.ErrNotFound, language)
	case codes.Unavailable:
		return encodeApplicationError(common.ErrServiceUnavailable, language)
	case codes.DeadlineExceeded:
		return encodeApplicationError(common.ErrDeadlineExceeded, language)
	default:
		return status.Error(
			codes.Internal,
			common.ErrInternalServerError.Message(language),
		)
	}
}

func encodeFieldErrors(
	fieldErrors dictionary.FieldErrors,
	language locale.Language,
) error {
	messages := fieldErrors.Messages(language)
	fields := make([]string, 0, len(messages))
	for field := range messages {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	violations := make([]*errdetails.BadRequest_FieldViolation, 0)
	for _, field := range fields {
		for _, message := range messages[field] {
			violations = append(violations, &errdetails.BadRequest_FieldViolation{
				Field:       field,
				Description: message,
			})
		}
	}

	base := status.New(
		codes.InvalidArgument,
		common.ErrBadRequest.Message(language),
	)
	withDetails, detailsErr := base.WithDetails(&errdetails.BadRequest{
		FieldViolations: violations,
	})
	if detailsErr != nil {
		return base.Err()
	}
	return withDetails.Err()
}

func encodeApplicationError(
	applicationError *dictionary.Error,
	language locale.Language,
) error {
	grpcCode, overridden := applicationError.GRPCCode()
	if !overridden {
		grpcCode = dictionary.GRPCCodeFromHTTPStatus(applicationError.HTTPStatus())
	}

	base := status.New(grpcCode, applicationError.Message(language))
	withDetails, detailsErr := base.WithDetails(&errdetails.ErrorInfo{
		Reason: applicationError.Key(),
		Domain: errorDomain,
		Metadata: map[string]string{
			"code":        applicationError.Code(),
			"http_status": strconv.Itoa(applicationError.HTTPStatus()),
		},
	})
	if detailsErr != nil {
		return base.Err()
	}
	return withDetails.Err()
}
