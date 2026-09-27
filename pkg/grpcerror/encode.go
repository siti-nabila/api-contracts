package grpcerror

import (
	"errors"
	"sort"
	"strconv"

	"github.com/siti-nabila/api-contracts/pkg/dictionary"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/common"
	"github.com/siti-nabila/api-contracts/pkg/locale"
	errorpackage "github.com/siti-nabila/error-package"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const errorDomain = "errors.siti-nabila.dev"

type Encoder struct {
	mappings []CodeMapping
}

// var defaultEncoder = NewEncoder(AuthCodeMappings()...)

func NewEncoder(mappings ...CodeMapping) *Encoder {
	return &Encoder{mappings: cloneCodeMappings(mappings)}
}

// func Encode(err error, language locale.Language) error {
// 	return defaultEncoder.Encode(err, language)
// }

func (encoder *Encoder) Encode(err error, language locale.Language) error {
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
		return encodeFieldErrors(map[string][]error(fieldErrors), language)
	}

	if fieldErrors, ok := errors.AsType[errorpackage.Errors](err); ok {
		return encodeFieldErrors(map[string][]error(fieldErrors), language)
	}

	if applicationError, ok := errors.AsType[*dictionary.Error](err); ok {
		grpcCode, exists := resolveCode(err, encoder.mappings)
		if !exists {
			return internalError(language)
		}
		return encodeApplicationError(applicationError, language, grpcCode)
	}

	return internalError(language)
}

func encodeGRPCStatus(code codes.Code, language locale.Language) error {
	switch code {
	case codes.InvalidArgument:
		return encodeApplicationError(common.ErrBadRequest, language, code)
	case codes.Unauthenticated:
		return encodeApplicationError(common.ErrUnauthorized, language, code)
	case codes.PermissionDenied:
		return encodeApplicationError(common.ErrForbidden, language, code)
	case codes.NotFound:
		return encodeApplicationError(common.ErrNotFound, language, code)
	case codes.Unavailable:
		return encodeApplicationError(common.ErrServiceUnavailable, language, code)
	case codes.DeadlineExceeded:
		return encodeApplicationError(common.ErrDeadlineExceeded, language, code)
	default:
		return internalError(language)
	}
}

func encodeFieldErrors(
	fieldErrors map[string][]error,
	language locale.Language,
) error {
	messages := make(map[string][]string)
	appendFieldMessages(messages, "", fieldErrors, language)
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

func appendFieldMessages(
	result map[string][]string,
	prefix string,
	fieldErrors map[string][]error,
	language locale.Language,
) {
	for field, errorList := range fieldErrors {
		fullField := field
		if prefix != "" {
			fullField = prefix + "." + field
		}

		for _, err := range errorList {
			switch typed := err.(type) {
			case dictionary.FieldErrors:
				appendFieldMessages(result, fullField, map[string][]error(typed), language)
			case errorpackage.Errors:
				appendFieldMessages(result, fullField, map[string][]error(typed), language)
			case interface{ Message(locale.Language) string }:
				result[fullField] = append(result[fullField], typed.Message(language))
			case interface {
				Message(errorpackage.LanguageCode) string
			}:
				result[fullField] = append(
					result[fullField],
					typed.Message(errorpackage.LanguageCode(language)),
				)
			case nil:
				result[fullField] = append(result[fullField], "<nil>")
			default:
				result[fullField] = append(result[fullField], err.Error())
			}
		}
	}
}

func encodeApplicationError(
	applicationError *dictionary.Error,
	language locale.Language,
	grpcCode codes.Code,
) error {
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

func internalError(language locale.Language) error {
	return status.Error(
		codes.Internal,
		common.ErrInternalServerError.Message(language),
	)
}
