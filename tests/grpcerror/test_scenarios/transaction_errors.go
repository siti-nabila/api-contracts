package test_scenarios

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/siti-nabila/api-contracts/pkg/dictionary"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/auth"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/common"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/user"
	"github.com/siti-nabila/api-contracts/pkg/grpcerror"
	"github.com/siti-nabila/api-contracts/pkg/grpcerror/mapping"
	"github.com/siti-nabila/api-contracts/pkg/locale"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func transactionErrors(t *testing.T) {
	mappings := append(mapping.CommonCodeMappings(), mapping.AuthCodeMappings()...)
	mappings = append(mappings, mapping.UserCodeMappings()...)
	encoder := grpcerror.NewEncoder(mappings...)
	if errors.Is(auth.ErrResourceBusy, user.ErrResourceBusy) {
		t.Fatal("auth and user resource busy must have distinct identities")
	}
	cases := []struct {
		name       string
		err        error
		want       *dictionary.Error
		code       codes.Code
		appCode    string
		httpStatus string
	}{
		{"user idempotency conflict", fmt.Errorf("private request: %w", user.ErrIdempotencyConflict), user.ErrIdempotencyConflict, codes.Aborted, "IC", "409"},
		{"auth busy", auth.ErrResourceBusy, auth.ErrResourceBusy, codes.Aborted, "RB", "409"},
		{"wrapped user busy", fmt.Errorf("private database detail: %w", user.ErrResourceBusy), user.ErrResourceBusy, codes.Aborted, "RB", "409"},
		{"dependency", fmt.Errorf("private database detail: %w", common.ErrDependencyUnavailable), common.ErrDependencyUnavailable, codes.Unavailable, "DU", "503"},
		{"legacy unavailable", common.ErrServiceUnavailable, common.ErrServiceUnavailable, codes.Unimplemented, "SU", "503"},
		{"cancellation catalog", common.ErrRequestCanceled, common.ErrRequestCanceled, codes.Canceled, "RC", "499"},
		{"wrapped cancellation", fmt.Errorf("private database detail: %w", context.Canceled), common.ErrRequestCanceled, codes.Canceled, "RC", "499"},
		{"wrapped deadline", fmt.Errorf("private database detail: %w", context.DeadlineExceeded), common.ErrDeadlineExceeded, codes.DeadlineExceeded, "TO", "504"},
		{"raw grpc canceled", status.Error(codes.Canceled, "private database detail"), common.ErrRequestCanceled, codes.Canceled, "RC", "499"},
		{"raw grpc deadline", status.Error(codes.DeadlineExceeded, "private database detail"), common.ErrDeadlineExceeded, codes.DeadlineExceeded, "TO", "504"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, lang := range []locale.Language{locale.English, locale.Indonesian, locale.Language("id-ID"), locale.Language("fr")} {
				encoded := encoder.Encode(tc.err, lang)
				got, ok := grpcerror.Decode(encoded)
				if !ok || got.Code != tc.code || got.Reason != tc.want.Key() || got.Message != tc.want.Message(lang) {
					t.Fatalf("locale %s: decoded = %#v, want %s / %s with localized safe message", lang, got, tc.code, tc.want.Key())
				}
				if got.Metadata["code"] != tc.appCode || got.Metadata["http_status"] != tc.httpStatus {
					t.Fatalf("metadata = %#v, want code=%s http_status=%s", got.Metadata, tc.appCode, tc.httpStatus)
				}
				if twice := encoder.Encode(encoded, lang); twice != encoded {
					t.Fatal("encoding detailed errors must preserve the original error")
				}
			}
		})
	}
	unknown := encoder.Encode(errors.New("private database detail"), locale.English)
	if status.Code(unknown) != codes.Internal || status.Convert(unknown).Message() != common.ErrInternalServerError.Message(locale.English) {
		t.Fatal("unknown errors must remain redacted internal failures")
	}
	wrapped := fmt.Errorf("private database detail: %w", auth.ErrResourceBusy)
	if !errors.Is(wrapped, auth.ErrResourceBusy) {
		t.Fatal("wrapping must preserve resource busy identity")
	}
	if _, ok := auth.Registry().Lookup("auth.resource_busy"); !ok {
		t.Fatal("auth busy catalog missing")
	}
	if _, ok := user.Registry().Lookup("user.resource_busy"); !ok {
		t.Fatal("user busy catalog missing")
	}
}
