package test_scenarios

import (
	userv1 "github.com/siti-nabila/api-contracts/pb/user/v1"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"testing"
)

func UserManagement() []testutils.Scenario {
	return []testutils.Scenario{
		{Name: "four exact unary user management methods", Run: managementMethods},
		{Name: "management presence mask and version survive the wire", Run: managementPresence},
		{Name: "legacy fields wrappers and additive user fields survive legacy wire hops", Run: managementCompatibility},
		{Name: "management requests expose no actor or privilege injection fields", Run: managementInputs},
		{Name: "safe management response fixtures carry no credentials", Run: managementSafeFixtures},
	}
}
func managementMethods(t *testing.T) {
	s := userv1.File_user_v1_user_service_proto.Services().Get(0)
	if s.Methods().Len() != 15 {
		t.Fatal("UserService inventory changed")
	}
	for name, full := range map[string]string{
		"GetUser":          userv1.UserService_GetUser_FullMethodName,
		"CreateUser":       userv1.UserService_CreateUser_FullMethodName,
		"UpdateUser":       userv1.UserService_UpdateUser_FullMethodName,
		"ChangeUserStatus": userv1.UserService_ChangeUserStatus_FullMethodName,
	} {
		m := s.Methods().ByName(protoreflect.Name(name))
		if m == nil || string(m.Input().FullName()) != "user.v1."+name+"Request" || string(m.Output().FullName()) != "user.v1."+name+"Response" || m.IsStreamingClient() || m.IsStreamingServer() || full != "/user.v1.UserService/"+name {
			t.Fatal("method contract changed", name)
		}
	}
}
func managementPresence(t *testing.T) {
	for _, v := range []*string{nil, proto.String(""), proto.String("ACTIVE")} {
		in := &userv1.CreateUserRequest{Status: v}
		out := new(userv1.CreateUserRequest)
		managementRoundTrip(t, in, out)
		if out.ProtoReflect().Has(out.ProtoReflect().Descriptor().Fields().ByName("status")) != (v != nil) {
			t.Fatal("status presence lost")
		}
	}
	for _, v := range []*uint64{nil, proto.Uint64(0), proto.Uint64(7)} {
		for _, value := range []*string{nil, proto.String(""), proto.String("Ana")} {
			in := &userv1.UpdateUserRequest{UserId: 42, ExpectedUserVersion: v, Name: value, Address: proto.String(""), FieldMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "address"}}}
			out := new(userv1.UpdateUserRequest)
			managementRoundTrip(t, in, out)
			for field, want := range map[protoreflect.Name]bool{"expected_user_version": v != nil, "name": value != nil, "address": true, "phone": false} {
				if out.ProtoReflect().Has(out.ProtoReflect().Descriptor().Fields().ByName(field)) != want {
					t.Fatal("field presence lost", field)
				}
			}
		}
		in := &userv1.ChangeUserStatusRequest{UserId: 42, ExpectedUserVersion: v, Status: proto.String("SUSPENDED"), Reason: "Policy review"}
		managementRoundTrip(t, in, new(userv1.ChangeUserStatusRequest))
	}
}
func managementRoundTrip(t *testing.T, in, out proto.Message) {
	t.Helper()
	wire, err := proto.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if err = proto.Unmarshal(wire, out); err != nil || !proto.Equal(in, out) {
		t.Fatal("wire roundtrip", err)
	}
}
func managementCompatibility(t *testing.T) {
	old, err := protodesc.NewFile(legacyFile(t, "user.payload.json"), protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []protoreflect.Name{"User", "CreateUserRequest", "CreateUserResponse", "GetUserRequest", "GetUserResponse"} {
		legacy := old.Messages().ByName(name)
		current := userv1.File_user_v1_user_payload_proto.Messages().ByName(name)
		for i := 0; i < legacy.Fields().Len(); i++ {
			f := legacy.Fields().Get(i)
			got := current.Fields().ByNumber(f.Number())
			if got == nil || got.Name() != f.Name() || got.Kind() != f.Kind() || got.Cardinality() != f.Cardinality() || (f.Message() != nil && got.Message().FullName() != f.Message().FullName()) {
				t.Fatal("legacy field changed", f.FullName())
			}
		}
	}
	want := &userv1.User{Id: 42, Name: "Ana", Email: "ana@example.test", PhoneNo: "123456", Address: "Test address", Status: "ACTIVE", Version: 1, CreatedAt: &timestamppb.Timestamp{Seconds: 12345}, MustChangePassword: true}
	wire, err := proto.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	legacy := dynamicpb.NewMessage(old.Messages().ByName("User"))
	if err = proto.Unmarshal(wire, legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.Get(legacy.Descriptor().Fields().ByNumber(4)).String() != "" || len(legacy.GetUnknown()) == 0 {
		t.Fatal("legacy credential or additive fields changed")
	}
	wire, err = proto.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	got := new(userv1.User)
	if err = proto.Unmarshal(wire, got); err != nil || !proto.Equal(want, got) {
		t.Fatal("legacy hop lost metadata", err)
	}
}
func managementInputs(t *testing.T) {
	for _, m := range []proto.Message{&userv1.CreateUserRequest{}, &userv1.GetUserRequest{}, &userv1.UpdateUserRequest{}, &userv1.ChangeUserStatusRequest{}} {
		fields := m.ProtoReflect().Descriptor().Fields()
		for i := 0; i < fields.Len(); i++ {
			name := string(fields.Get(i).Name())
			for _, bad := range []string{"actor", "role", "permission", "application", "auth_version", "must_change_password", "hash"} {
				if strings.Contains(name, bad) {
					t.Error("unexpected caller controlled field", name)
				}
			}
			if name == "password" && m.ProtoReflect().Descriptor().Name() != "CreateUserRequest" {
				t.Error("unexpected credential input")
			}
		}
	}
}
func managementSafeFixtures(t *testing.T) {
	u := &userv1.User{Id: 42, Email: "ana@example.test", Status: "ACTIVE", Version: 1, MustChangePassword: true}
	for _, m := range []proto.Message{&userv1.GetUserResponse{User: u}, &userv1.CreateUserResponse{User: u}, &userv1.UpdateUserResponse{User: u, Changed: true}, &userv1.ChangeUserStatusResponse{UserId: 42, Status: "SUSPENDED", Version: 2, Changed: true}} {
		b, err := protojson.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		for _, bad := range []string{`"password":`, `"token":`, `"refreshToken":`, `"authVersion":`, `"hash":`} {
			if strings.Contains(string(b), bad) {
				t.Fatal("unsafe fixture serialization", string(b))
			}
		}
	}
	// User.password is a legacy compatibility field. Runtime safe mapping is
	// verified by grpc-auth registered-response tests, not claimed here.
	fields := (&userv1.User{}).ProtoReflect().Descriptor().Fields()
	for name, num := range map[protoreflect.Name]protoreflect.FieldNumber{"address": 7, "status": 8, "version": 9, "created_at": 10, "must_change_password": 11, "roles": 12, "application_access": 13} {
		if f := fields.ByName(name); f == nil || f.Number() != num {
			t.Fatal("safe metadata tag changed", name)
		}
	}
}
