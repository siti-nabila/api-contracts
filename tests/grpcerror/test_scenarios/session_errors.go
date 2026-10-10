package test_scenarios

import (
	"errors"
	"fmt"
	"github.com/siti-nabila/api-contracts/pkg/dictionary"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/auth"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/common"
	"github.com/siti-nabila/api-contracts/pkg/grpcerror"
	"github.com/siti-nabila/api-contracts/pkg/locale"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"testing"
	"time"
)

func sessionErrors(t *testing.T) {
	cases := []struct {
		err             *dictionary.Error
		key, code, http string
		grpc            codes.Code
	}{
		{auth.ErrServiceAuthenticationFailed, "auth.service_authentication_failed", "SF", "401", codes.Unauthenticated},
		{auth.ErrInvalidSession, "auth.invalid_session", "SI", "401", codes.Unauthenticated},
		{auth.ErrLoginFailed, "auth.login_failed", "LF", "401", codes.Unauthenticated},
		{common.ErrRateLimited, "common.rate_limited", "RL", "429", codes.ResourceExhausted},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			for _, lang := range []locale.Language{locale.English, locale.Indonesian, locale.Language("id-ID"), locale.Language("fr")} {
				wrapped := fmt.Errorf("private credential/database detail: %w", tc.err)
				if !errors.Is(wrapped, tc.err) {
					t.Fatal("identity lost")
				}
				got, ok := grpcerror.Decode(newEncoder().Encode(wrapped, lang))
				if !ok || got.Code != tc.grpc || got.Reason != tc.key || got.Metadata["code"] != tc.code || got.Metadata["http_status"] != tc.http || got.Message != tc.err.Message(lang) || got.RetryInfo != nil {
					t.Fatalf("locale %s: %#v", lang, got)
				}
				if got.Message == "" {
					t.Fatal("empty localization")
				}
				if lang == locale.Language("id-ID") && got.Message != tc.err.Message(locale.Indonesian) {
					t.Fatal("regional fallback changed")
				}
				if lang == locale.Language("fr") && got.Message != tc.err.Message(locale.English) {
					t.Fatal("English fallback changed")
				}
			}
		})
	}
	// A generic login failure cannot silently remap the old validation error.
	got, _ := grpcerror.Decode(newEncoder().Encode(auth.ErrPasswordMismatch, locale.English))
	if got.Code != codes.InvalidArgument || got.Reason != "auth.password_mismatch" {
		t.Fatal("legacy password mismatch changed")
	}
}

func rateRetryDetails(t *testing.T) {
	encoder := newEncoder()
	encoded := encoder.Encode(common.ErrRateLimited, locale.Indonesian)
	base := status.Convert(encoded)
	info := new(errdetails.ErrorInfo)
	if len(base.Details()) != 1 {
		t.Fatal("catalog error detail missing")
	}
	proto.Merge(info, base.Details()[0].(*errdetails.ErrorInfo))
	for _, reason := range []string{"quota", "capacity"} {
		info.Metadata["limit_reason"] = reason
		retry := &errdetails.RetryInfo{RetryDelay: durationpb.New(3 * time.Second)}
		// Build one ErrorInfo + RetryInfo; do not append a second conflicting reason.
		detailed, err := status.New(base.Code(), base.Message()).WithDetails(info, retry)
		if err != nil {
			t.Fatal(err)
		}
		wire := detailed.Err()
		if encoder.Encode(wire, locale.English) != wire {
			t.Fatal("detailed retry error reencoded/lost")
		}
		got, ok := grpcerror.Decode(wire)
		if !ok || got.Code != codes.ResourceExhausted || got.Reason != "common.rate_limited" || got.Metadata["limit_reason"] != reason || got.Metadata["code"] != "RL" || got.Metadata["http_status"] != "429" || got.Message != common.ErrRateLimited.Message(locale.Indonesian) || !proto.Equal(got.RetryInfo, retry) {
			t.Fatalf("retry detail lost: %#v", got)
		}
	}
}
