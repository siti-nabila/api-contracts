package test_scenarios

import (
	authv1 "github.com/siti-nabila/api-contracts/pb/auth/v1"
	sessionv1 "github.com/siti-nabila/api-contracts/pb/session/v1"
	userv1 "github.com/siti-nabila/api-contracts/pb/user/v1"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"testing"
)

func Lifecycle() []testutils.Scenario {
	return []testutils.Scenario{
		{Name: "user list creation timestamp is additive and survives a legacy wire hop", Run: userListCreatedAt},
		{Name: "nine unary lifecycle methods retain exact names and message types", Run: lifecycleMethods},
		{Name: "credential mode and expected version retain omitted versus empty presence", Run: lifecyclePresence},
		{Name: "lifecycle management outputs recursively exclude credential fields", Run: lifecycleSafeOutputs},
	}
}

func lifecycleMethods(t *testing.T) {
	cases := []struct {
		service                     protoreflect.ServiceDescriptor
		name, input, output, method string
	}{
		{authv1.File_auth_v1_auth_service_proto.Services().Get(0), "RefreshSession", "auth.v1.RefreshSessionRequest", "user.v1.UserTokenResponse", authv1.AuthService_RefreshSession_FullMethodName},
		{authv1.File_auth_v1_auth_service_proto.Services().Get(0), "Logout", "auth.v1.LogoutRequest", "auth.v1.LogoutResponse", authv1.AuthService_Logout_FullMethodName},
		{authv1.File_auth_v1_auth_service_proto.Services().Get(0), "ChangePassword", "auth.v1.ChangePasswordRequest", "auth.v1.ChangePasswordResponse", authv1.AuthService_ChangePassword_FullMethodName},
		{sessionv1.File_session_v1_session_service_proto.Services().Get(0), "ListMySessions", "session.v1.ListMySessionsRequest", "session.v1.ListSessionsResponse", sessionv1.SessionService_ListMySessions_FullMethodName},
		{sessionv1.File_session_v1_session_service_proto.Services().Get(0), "RevokeMySession", "session.v1.RevokeMySessionRequest", "session.v1.RevokeSessionResponse", sessionv1.SessionService_RevokeMySession_FullMethodName},
		{sessionv1.File_session_v1_session_service_proto.Services().Get(0), "RevokeAllMySessions", "google.protobuf.Empty", "session.v1.RevokeAllMySessionsResponse", sessionv1.SessionService_RevokeAllMySessions_FullMethodName},
		{sessionv1.File_session_v1_session_service_proto.Services().Get(0), "ListUserSessions", "session.v1.ListUserSessionsRequest", "session.v1.ListSessionsResponse", sessionv1.SessionService_ListUserSessions_FullMethodName},
		{sessionv1.File_session_v1_session_service_proto.Services().Get(0), "RevokeUserSession", "session.v1.RevokeUserSessionRequest", "session.v1.RevokeSessionResponse", sessionv1.SessionService_RevokeUserSession_FullMethodName},
		{sessionv1.File_session_v1_session_service_proto.Services().Get(0), "RevokeAllUserSessions", "session.v1.RevokeAllUserSessionsRequest", "session.v1.RevokeAllUserSessionsResponse", sessionv1.SessionService_RevokeAllUserSessions_FullMethodName},
	}
	if cases[0].service.Methods().Len() != 3 || cases[3].service.Methods().Len() != 6 {
		t.Fatal("unexpected lifecycle RPC inventory")
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.service.Methods().ByName(protoreflect.Name(tc.name))
			if m == nil || string(m.Input().FullName()) != tc.input || string(m.Output().FullName()) != tc.output || m.IsStreamingClient() || m.IsStreamingServer() {
				t.Fatal("RPC contract changed")
			}
			if tc.method != "/"+string(tc.service.FullName())+"/"+tc.name {
				t.Fatal("full method constant changed")
			}
		})
	}
	// Refresh deliberately returns the existing Login message, preserving its tags.
	fields := (&userv1.UserTokenResponse{}).ProtoReflect().Descriptor().Fields()
	for name, num := range map[string]protoreflect.FieldNumber{"token": 1, "refresh_token": 4, "session_id": 5, "access_expires_at": 6, "session_expires_at": 7, "must_change_password": 8} {
		f := fields.ByName(protoreflect.Name(name))
		if f == nil || f.Number() != num {
			t.Fatal("shared credential field changed", name)
		}
	}
}

func lifecyclePresence(t *testing.T) {
	empty := ""
	token := "synthetic-wire-value"
	for _, value := range []*string{nil, &empty, &token} {
		in := &authv1.LogoutRequest{RefreshToken: value}
		wire, err := proto.Marshal(in)
		if err != nil {
			t.Fatal(err)
		}
		out := new(authv1.LogoutRequest)
		if err = proto.Unmarshal(wire, out); err != nil {
			t.Fatal(err)
		}
		f := out.ProtoReflect().Descriptor().Fields().ByName("refresh_token")
		if out.ProtoReflect().Has(f) != (value != nil) || !proto.Equal(in, out) {
			t.Fatal("Logout omitted/present-empty credential mode lost")
		}
	}
	zero, positive := uint64(0), uint64(8)
	for _, version := range []*uint64{nil, &zero, &positive} {
		in := &sessionv1.RevokeAllUserSessionsRequest{UserId: 42, ExpectedUserVersion: version}
		wire, err := proto.Marshal(in)
		if err != nil {
			t.Fatal(err)
		}
		out := new(sessionv1.RevokeAllUserSessionsRequest)
		if err = proto.Unmarshal(wire, out); err != nil {
			t.Fatal(err)
		}
		f := out.ProtoReflect().Descriptor().Fields().ByName("expected_user_version")
		if f.Number() != 2 || f.Kind() != protoreflect.Uint64Kind || out.ProtoReflect().Has(f) != (version != nil) || !proto.Equal(in, out) {
			t.Fatal("expected entity version presence/type lost")
		}
	}
}

func lifecycleSafeOutputs(t *testing.T) {
	seen := map[protoreflect.FullName]bool{}
	var check func(protoreflect.MessageDescriptor)
	check = func(d protoreflect.MessageDescriptor) {
		if seen[d.FullName()] {
			return
		}
		seen[d.FullName()] = true
		for i := 0; i < d.Fields().Len(); i++ {
			f := d.Fields().Get(i)
			for _, bad := range []string{"password", "token", "hash", "secret", "auth_version"} {
				if strings.Contains(string(f.Name()), bad) {
					t.Errorf("unsafe lifecycle output field %s", f.FullName())
				}
			}
			if f.Message() != nil {
				check(f.Message())
			}
		}
	}
	for _, m := range []proto.Message{&authv1.LogoutResponse{}, &authv1.ChangePasswordResponse{}, &sessionv1.ListSessionsResponse{}, &sessionv1.RevokeSessionResponse{}, &sessionv1.RevokeAllMySessionsResponse{}, &sessionv1.RevokeAllUserSessionsResponse{}} {
		check(m.ProtoReflect().Descriptor())
	}
	in := &sessionv1.ListSessionsResponse{Items: []*sessionv1.Session{{Id: "00000000-0000-4000-8000-000000000001", UserId: 42, Status: "ACTIVE", IsCurrent: true}}, NextCursor: "safe-cursor"}
	wire, err := proto.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	out := new(sessionv1.ListSessionsResponse)
	if err = proto.Unmarshal(wire, out); err != nil || !proto.Equal(in, out) {
		t.Fatal("safe session metadata roundtrip failed", err)
	}
}

func userListCreatedAt(t *testing.T) {
	fields := (&userv1.UserListItem{}).ProtoReflect().Descriptor().Fields()
	for name, num := range map[string]protoreflect.FieldNumber{"id": 1, "email": 2, "name": 3, "address": 4, "phone": 5, "status": 6, "version": 7, "roles": 8, "application_access": 9, "created_at": 10} {
		f := fields.ByName(protoreflect.Name(name))
		if f == nil || f.Number() != num {
			t.Fatal("list field tag changed", name)
		}
	}
	f := fields.ByName("created_at")
	if f.Kind() != protoreflect.MessageKind || f.Message().FullName() != "google.protobuf.Timestamp" {
		t.Fatal("created_at type changed")
	}
	want := &userv1.UserListItem{Id: 42, Email: "ana@example.test", Name: "Ana", Status: "ACTIVE", Version: 3, CreatedAt: &timestamppb.Timestamp{Seconds: 12345}}
	wire, err := proto.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	old, err := protodesc.NewFile(legacyFile(t, "user.payload.json"), protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}
	legacy := dynamicpb.NewMessage(old.Messages().ByName("UserListItem"))
	if err = proto.Unmarshal(wire, legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.Get(legacy.Descriptor().Fields().ByNumber(1)).Int() != 42 || len(legacy.GetUnknown()) == 0 {
		t.Fatal("legacy ID or additive fields lost")
	}
	wire, err = proto.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	got := new(userv1.UserListItem)
	if err = proto.Unmarshal(wire, got); err != nil || !proto.Equal(want, got) {
		t.Fatal("list timestamp did not survive legacy wire hop", err)
	}
}
