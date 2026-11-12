package test_scenarios

import (
	"errors"
	"fmt"
	"github.com/siti-nabila/api-contracts/pkg/dictionary"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/auth"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/user"
	"github.com/siti-nabila/api-contracts/pkg/grpcerror"
	grpcmapping "github.com/siti-nabila/api-contracts/pkg/grpcerror/mapping"
	"github.com/siti-nabila/api-contracts/pkg/locale"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"google.golang.org/grpc/codes"
	"strings"
	"testing"
)

func LifecycleErrors() []testutils.Scenario {
	return []testutils.Scenario{{Name: "lifecycle errors preserve qualified identity locale metadata and redaction", Run: lifecycleErrors}}
}

func lifecycleErrors(t *testing.T) {
	mappings := append(grpcmapping.CommonCodeMappings(), grpcmapping.AuthCodeMappings()...)
	mappings = append(mappings, grpcmapping.UserCodeMappings()...)
	encoder := grpcerror.NewEncoder(mappings...)
	cases := []struct {
		err                     *dictionary.Error
		key, code, http, en, id string
		grpc                    codes.Code
	}{
		{user.ErrInvalidStatusTransition, "user.invalid_status_transition", "IST", "409", "User status transition is not allowed.", "Perubahan status pengguna tidak diizinkan.", codes.FailedPrecondition},
		{user.ErrSelfInactivationForbidden, "user.self_inactivation_forbidden", "SIF", "403", "You cannot suspend or deactivate your own account.", "Anda tidak dapat menangguhkan atau menonaktifkan akun sendiri.", codes.PermissionDenied},
		{user.ErrLastAdminRequired, "user.last_admin_required", "LAR", "409", "At least one active platform administrator is required.", "Minimal satu administrator platform aktif harus tersedia.", codes.FailedPrecondition},
		{auth.ErrRefreshTokenReused, "auth.refresh_token_reused", "RR", "401", "Refresh token was already used. Log in again.", "Token refresh sudah digunakan. Silakan login kembali.", codes.Unauthenticated},
		{user.ErrVersionConflict, "user.version_conflict", "VC", "409", "User version has changed. Reload and retry.", "Versi pengguna telah berubah. Muat ulang dan coba kembali.", codes.Aborted},
		{user.ErrIdempotencyConflict, "user.idempotency_conflict", "IC", "409", "Idempotency key was already used with a different request.", "Kunci idempotensi sudah digunakan untuk permintaan yang berbeda.", codes.Aborted},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			wrapped := fmt.Errorf("private-digest-detail: %w", tc.err)
			if !errors.Is(wrapped, tc.err) {
				t.Fatal("error identity lost")
			}
			for _, lang := range []locale.Language{locale.English, locale.Indonesian, locale.Language("id-ID"), locale.Language("fr")} {
				got, ok := grpcerror.Decode(encoder.Encode(wrapped, lang))
				want := tc.en
				if lang == locale.Indonesian || lang == locale.Language("id-ID") {
					want = tc.id
				}
				if !ok || got.Code != tc.grpc || got.Reason != tc.key || got.Metadata["code"] != tc.code || got.Metadata["http_status"] != tc.http || got.Message != want || strings.Contains(got.Message, "private-digest-detail") {
					t.Fatalf("bad localized wire error: %#v", got)
				}
			}
		})
	}
	got, _ := grpcerror.Decode(encoder.Encode(auth.ErrPasswordMismatch, locale.English))
	if got.Code != codes.InvalidArgument || got.Reason != "auth.password_mismatch" {
		t.Fatal("legacy password mismatch remapped")
	}
}
