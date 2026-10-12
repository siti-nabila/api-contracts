package test_scenarios

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/siti-nabila/api-contracts/pkg/dictionary"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/application"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/common"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/role"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/user"
	"github.com/siti-nabila/api-contracts/pkg/grpcerror"
	"github.com/siti-nabila/api-contracts/pkg/grpcerror/mapping"
	"github.com/siti-nabila/api-contracts/pkg/locale"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RoleErrors() []testutils.Scenario {
	return []testutils.Scenario{{Name: "role and application error identity locale cause and redaction", Run: roleErrors}}
}

func roleErrors(t *testing.T) {
	maps := append(mapping.CommonCodeMappings(), mapping.RoleCodeMappings()...)
	maps = append(maps, mapping.ApplicationCodeMappings()...)
	maps = append(maps, mapping.UserCodeMappings()...)
	encoder := grpcerror.NewEncoder(maps...)
	cases := []struct {
		err                   *dictionary.Error
		code                  codes.Code
		appCode, http, en, id string
	}{
		{role.ErrRoleNotFound, codes.NotFound, "RNF", "404", "Role not found.", "Role tidak ditemukan."},
		{role.ErrRoleCodeAlreadyExists, codes.AlreadyExists, "RCE", "409", "Role code already exists.", "Kode role sudah ada."},
		{role.ErrSystemRoleProtected, codes.FailedPrecondition, "SRP", "409", "System roles cannot be modified.", "Role sistem tidak dapat diubah."},
		{role.ErrInvalidPermissionSet, codes.InvalidArgument, "IPS", "400", "Permissions must be registered in the role application.", "Permission harus terdaftar pada aplikasi role."},
		{role.ErrVersionConflict, codes.Aborted, "VC", "409", "Role version has changed. Reload and retry.", "Versi role telah berubah. Muat ulang dan coba kembali."},
		{role.ErrIdempotencyConflict, codes.Aborted, "IC", "409", "Idempotency key was already used with a different request.", "Kunci idempotensi sudah digunakan untuk permintaan yang berbeda."},
		{role.ErrResourceBusy, codes.Aborted, "RB", "409", "Resource is busy. Retry the request.", "Resource sedang digunakan. Ulangi permintaan."},
		{application.ErrApplicationNotFound, codes.NotFound, "ANF", "404", "Application not found.", "Aplikasi tidak ditemukan."},
		{application.ErrApplicationInactive, codes.FailedPrecondition, "AI", "409", "Application is not active.", "Aplikasi tidak aktif."},
		{application.ErrSystemApplicationProtected, codes.FailedPrecondition, "SAP", "409", "System applications cannot contain custom roles.", "Aplikasi sistem tidak dapat memiliki role kustom."},
		{user.ErrGrantTargetInactive, codes.FailedPrecondition, "GTI", "409", "Grant target must be an active user.", "Target pemberian role harus pengguna aktif."},
		{user.ErrSelfAdminRoleRevokeForbidden, codes.PermissionDenied, "SARF", "403", "You cannot revoke your own platform administrator role.", "Anda tidak dapat mencabut role administrator platform sendiri."},
		{role.ErrRoleDisabled, codes.FailedPrecondition, "RD", "409", "Role is disabled.", "Role dinonaktifkan."},
		{application.ErrApplicationAccessRequired, codes.FailedPrecondition, "AAR", "409", "Active application access is required.", "Akses aplikasi aktif diperlukan."},
	}
	cause := errors.New("private database constraint detail")
	for _, c := range cases {
		t.Run(c.err.Key(), func(t *testing.T) {
			wrapped := fmt.Errorf("private target: %w", errors.Join(c.err, cause))
			if !errors.Is(wrapped, c.err) || !errors.Is(wrapped, cause) {
				t.Fatal("source identity/cause lost")
			}
			for lang, want := range map[locale.Language]string{locale.English: c.en, locale.Indonesian: c.id, "id-ID": c.id, "fr": c.en} {
				encoded := encoder.Encode(wrapped, lang)
				got, ok := grpcerror.Decode(encoded)
				if !ok || got.Code != c.code || got.Reason != c.err.Key() || got.Message != want || got.Metadata["code"] != c.appCode || got.Metadata["http_status"] != c.http {
					t.Fatalf("locale %s: %#v", lang, got)
				}
				def, found := got.Definition(role.Registry(), application.Registry(), user.Registry())
				if !found || def.Key() != c.err.Key() || def.Message(lang) != want {
					t.Fatal("decoded catalog definition missing")
				}
				if strings.Contains(encoded.Error(), "private") || strings.Contains(encoded.Error(), "constraint") {
					t.Fatal("private cause leaked")
				}
				if encoder.Encode(encoded, lang) != encoded {
					t.Fatal("detailed error re-encoded")
				}
			}
		})
	}
	if errors.Is(role.ErrResourceBusy, user.ErrResourceBusy) || errors.Is(role.ErrVersionConflict, user.ErrVersionConflict) {
		t.Fatal("feature identity collision")
	}
	if status.Code(encoder.Encode(common.ErrServiceUnavailable, locale.English)) != codes.Unimplemented {
		t.Fatal("legacy unavailable remapped")
	}
	if status.Code(encoder.Encode(user.ErrVersionConflict, locale.English)) != codes.Aborted {
		t.Fatal("legacy user mapping changed")
	}
	got := encoder.Encode(cause, locale.English)
	if status.Code(got) != codes.Internal || status.Convert(got).Message() != common.ErrInternalServerError.Message(locale.English) {
		t.Fatal("unknown error redaction changed")
	}
	// Malformed/presence business validation reuses the existing field boundary.
	invalid := dictionary.FieldErrors{"expected_role_version": []error{common.ErrBadRequest}}
	decoded, ok := grpcerror.Decode(encoder.Encode(invalid, locale.English))
	if !ok || decoded.Code != codes.InvalidArgument || len(decoded.FieldErrors["expected_role_version"]) != 1 {
		t.Fatal("field validation boundary changed")
	}
}
