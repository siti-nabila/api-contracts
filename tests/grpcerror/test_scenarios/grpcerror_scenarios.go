package test_scenarios

import (
	"errors"
	"testing"

	"github.com/siti-nabila/api-contracts/pkg/dictionary"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/auth"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/common"
	"github.com/siti-nabila/api-contracts/pkg/grpcerror"
	"github.com/siti-nabila/api-contracts/pkg/locale"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"google.golang.org/grpc/codes"
)

func All() []testutils.Scenario {
	return []testutils.Scenario{
		{
			Name: "encodes business error using derived grpc code",
			Run:  encodeDerivedCode,
		},
		{
			Name: "encodes business error using optional grpc override",
			Run:  encodeOverrideCode,
		},
		{
			Name: "encodes localized bad request field details",
			Run:  encodeBadRequest,
		},
		{
			Name: "hides unknown internal error",
			Run:  hideUnknownError,
		},
	}
}

func encodeDerivedCode(t *testing.T) {
	encoded := grpcerror.Encode(auth.ErrNotFound, locale.Indonesian)
	decoded, ok := grpcerror.Decode(encoded)
	if !ok {
		t.Fatal("Decode() ok = false, want true")
	}
	if decoded.Code != codes.NotFound {
		t.Errorf("decoded code = %s, want %s", decoded.Code, codes.NotFound)
	}
	if decoded.Reason != auth.ErrNotFound.Key() {
		t.Errorf("decoded reason = %q, want %q", decoded.Reason, auth.ErrNotFound.Key())
	}
	if decoded.Message != auth.ErrNotFound.Message(locale.Indonesian) {
		t.Errorf("decoded message = %q, want Indonesian dictionary message", decoded.Message)
	}
}

func encodeOverrideCode(t *testing.T) {
	encoded := grpcerror.Encode(auth.ErrPasswordMismatch, locale.English)
	decoded, ok := grpcerror.Decode(encoded)
	if !ok {
		t.Fatal("Decode() ok = false, want true")
	}
	if decoded.Code != codes.FailedPrecondition {
		t.Errorf(
			"decoded code = %s, want %s",
			decoded.Code,
			codes.FailedPrecondition,
		)
	}
}

func encodeBadRequest(t *testing.T) {
	fieldErrors := dictionary.FieldErrors{}
	fieldErrors.Add("email", auth.ErrRequired)
	fieldErrors.Add("email", auth.ErrMinLength(6))

	encoded := grpcerror.Encode(fieldErrors, locale.Indonesian)
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

func hideUnknownError(t *testing.T) {
	encoded := grpcerror.Encode(errors.New("database password leaked"), locale.English)
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
