package test_scenarios

import (
	"errors"
	"fmt"
	"testing"

	"github.com/siti-nabila/api-contracts/pkg/dictionary"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/auth"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/common"
	"github.com/siti-nabila/api-contracts/pkg/grpcerror"
	grpcmapping "github.com/siti-nabila/api-contracts/pkg/grpcerror/mapping"
	"github.com/siti-nabila/api-contracts/pkg/locale"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	errorpackage "github.com/siti-nabila/error-package"
	"google.golang.org/grpc/codes"
)

func All() []testutils.Scenario {
	return []testutils.Scenario{
		{
			Name: "encodes auth errors using explicit grpc mappings",
			Run:  encodeMappedCodes,
		},
		{
			Name: "encodes common errors using explicit grpc mappings",
			Run:  encodeCommonMappedCodes,
		},
		{
			Name: "encodes localized bad request field details",
			Run:  encodeBadRequest,
		},
		{
			Name: "encodes multiple field messages in an arbitrary language",
			Run:  encodeArbitraryLanguageFieldErrors,
		},
		{
			Name: "encodes error-package errors as bad request field details",
			Run:  encodeErrorPackageErrors,
		},
		{
			Name: "hides unmapped dictionary error as internal error",
			Run:  hideUnmappedDictionaryError,
		},
		{
			Name: "hides unknown internal error",
			Run:  hideUnknownError,
		},
	}
}

func encodeMappedCodes(t *testing.T) {
	encoder := newEncoder()
	testCases := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "minimum length", err: auth.ErrMinLength(6), code: codes.InvalidArgument},
		{name: "maximum length", err: auth.ErrMaxLength(50), code: codes.InvalidArgument},
		{name: "alpha only", err: auth.ErrAlphaOnly, code: codes.InvalidArgument},
		{name: "invalid email", err: auth.ErrInvalidEmail, code: codes.InvalidArgument},
		{name: "required", err: auth.ErrRequired, code: codes.InvalidArgument},
		{name: "wrapped password mismatch", err: fmt.Errorf("login: %w", auth.ErrPasswordMismatch), code: codes.InvalidArgument},
		{name: "not found", err: auth.ErrNotFound, code: codes.NotFound},
		{name: "already exists", err: auth.ErrDataExists, code: codes.AlreadyExists},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			applicationError, ok := errors.AsType[*dictionary.Error](testCase.err)
			if !ok {
				t.Fatalf("errors.AsType() ok = false for %T", testCase.err)
			}
			encoded := encoder.Encode(testCase.err, locale.Indonesian)
			decoded, ok := grpcerror.Decode(encoded)
			if !ok {
				t.Fatal("Decode() ok = false, want true")
			}
			if decoded.Code != testCase.code {
				t.Errorf("decoded code = %s, want %s", decoded.Code, testCase.code)
			}
			if decoded.Reason != applicationError.Key() {
				t.Errorf("decoded reason = %q, want %q", decoded.Reason, applicationError.Key())
			}
			if decoded.Message != applicationError.Message(locale.Indonesian) {
				t.Errorf("decoded message = %q, want Indonesian dictionary message", decoded.Message)
			}
		})
	}
}

func encodeCommonMappedCodes(t *testing.T) {
	encoder := newEncoder()
	testCases := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "bad request", err: common.ErrBadRequest, code: codes.InvalidArgument},
		{name: "not found", err: common.ErrNotFound, code: codes.NotFound},
		{name: "endpoint not found", err: common.ErrEndpointNotFound, code: codes.Unimplemented},
		{name: "service unavailable", err: common.ErrServiceUnavailable, code: codes.Unimplemented},
		{name: "unauthorized", err: common.ErrUnauthorized, code: codes.Unauthenticated},
		{name: "forbidden", err: common.ErrForbidden, code: codes.PermissionDenied},
		{name: "deadline exceeded", err: common.ErrDeadlineExceeded, code: codes.DeadlineExceeded},
		{name: "internal", err: common.ErrInternalServerError, code: codes.Internal},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			encoded := encoder.Encode(testCase.err, locale.English)
			decoded, ok := grpcerror.Decode(encoded)
			if !ok {
				t.Fatal("Decode() ok = false, want true")
			}
			if decoded.Code != testCase.code {
				t.Errorf("decoded code = %s, want %s", decoded.Code, testCase.code)
			}
		})
	}
}

func encodeBadRequest(t *testing.T) {
	encoder := newEncoder()
	fieldErrors := dictionary.FieldErrors{}
	fieldErrors.Add("email", auth.ErrRequired)
	fieldErrors.Add("email", auth.ErrMinLength(6))

	encoded := encoder.Encode(fieldErrors, locale.Indonesian)
	decoded, ok := grpcerror.Decode(encoded)
	if !ok {
		t.Fatal("Decode() ok = false, want true")
	}
	if decoded.Code != codes.InvalidArgument {
		t.Errorf("decoded code = %s, want %s", decoded.Code, codes.InvalidArgument)
	}
	want := []string{"harus diisi", "minimal panjang karakter adalah 6"}
	got := decoded.FieldErrors["email"]
	if len(got) != len(want) {
		t.Fatalf("field errors = %#v, want %#v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("field error[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}

func encodeArbitraryLanguageFieldErrors(t *testing.T) {
	encoder := newEncoder()
	fieldErrors := dictionary.FieldErrors{}
	fieldErrors.Add("email", auth.ErrRequired)
	fieldErrors.Add("email", auth.ErrMinLength(6))

	encoded := encoder.Encode(
		fmt.Errorf("validate request: %w", fieldErrors),
		locale.Language("zh-CN"),
	)
	decoded, ok := grpcerror.Decode(encoded)
	if !ok {
		t.Fatal("Decode() ok = false, want true")
	}
	want := []string{"必填", "最小长度为 6"}
	got := decoded.FieldErrors["email"]
	if len(got) != len(want) {
		t.Fatalf("field errors = %#v, want %#v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("field error[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}

func encodeErrorPackageErrors(t *testing.T) {
	encoder := newEncoder()
	fieldErrors := errorpackage.NewErrors()
	fieldErrors.Add("email", auth.ErrRequired)
	fieldErrors.Add("email", auth.ErrMinLength(6))

	encoded := encoder.Encode(fieldErrors, locale.Language("zh-CN"))
	decoded, ok := grpcerror.Decode(encoded)
	if !ok {
		t.Fatal("Decode() ok = false, want true")
	}
	if decoded.Code != codes.InvalidArgument {
		t.Errorf("decoded code = %s, want %s", decoded.Code, codes.InvalidArgument)
	}
	want := []string{"必填", "最小长度为 6"}
	got := decoded.FieldErrors["email"]
	if len(got) != len(want) {
		t.Fatalf("field errors = %#v, want %#v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("field error[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}

func hideUnmappedDictionaryError(t *testing.T) {
	encoded := newEncoder().Encode(auth.ErrGeneratingToken, locale.English)
	decoded, ok := grpcerror.Decode(encoded)
	if !ok {
		t.Fatal("Decode() ok = false, want true")
	}
	if decoded.Code != codes.Internal {
		t.Errorf("decoded code = %s, want %s", decoded.Code, codes.Internal)
	}
	if decoded.Message != common.ErrInternalServerError.Message(locale.English) {
		t.Errorf("decoded message = %q, want generic internal message", decoded.Message)
	}
	if decoded.Reason != "" {
		t.Errorf("decoded reason = %q, want empty reason", decoded.Reason)
	}
}

func hideUnknownError(t *testing.T) {
	encoded := newEncoder().Encode(
		errors.New("database password leaked"),
		locale.English,
	)
	decoded, ok := grpcerror.Decode(encoded)
	if !ok {
		t.Fatal("Decode() ok = false, want true")
	}
	if decoded.Code != codes.Internal {
		t.Errorf("decoded code = %s, want %s", decoded.Code, codes.Internal)
	}
	if decoded.Message != common.ErrInternalServerError.Message(locale.English) {
		t.Errorf("decoded message = %q, want generic internal message", decoded.Message)
	}
}

func newEncoder() *grpcerror.Encoder {
	mappings := append(
		grpcmapping.CommonCodeMappings(),
		grpcmapping.AuthCodeMappings()...,
	)
	return grpcerror.NewEncoder(mappings...)
}
