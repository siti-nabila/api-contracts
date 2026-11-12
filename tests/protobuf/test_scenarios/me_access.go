package test_scenarios

import (
	"testing"

	userv1 "github.com/siti-nabila/api-contracts/pb/user/v1"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

func MeAccess() []testutils.Scenario {
	return []testutils.Scenario{
		{Name: "Me keeps empty input and all fourteen legacy fields with one map at tag fifteen", Run: meAccessFields},
		{Name: "Me permission maps survive binary and proto JSON including empty maps", Run: meAccessWire},
		{Name: "legacy Me descriptors preserve unknown permission map on a wire hop", Run: meAccessLegacy},
		{Name: "Me rejects undeclared credential and session response fields", Run: meAccessRedaction},
	}
}

func meAccessFields(t *testing.T) {
	d := (&userv1.UserData{}).ProtoReflect().Descriptor()
	fields := []string{
		"id:uint64", "email:string", "fullname:string", "address:string", "phone:string",
		"role_codes:int32*", "role_names:string*", "created_at:google.protobuf.Timestamp",
		"status:string", "version:uint64", "roles:user.v1.UserRoleAssignment*",
		"application_access:user.v1.UserApplicationAccess*", "must_change_password:bool", "country_code:string",
	}
	if d.Fields().Len() != 15 {
		t.Fatal("Me field count", d.Fields().Len())
	}
	for i, field := range fields {
		assertRoleField(t, d, i, field)
	}
	f := d.Fields().ByNumber(15)
	if f == nil || f.Name() != "permissions" || !f.IsMap() || f.MapKey().Kind() != protoreflect.StringKind || f.MapValue().Kind() != protoreflect.BoolKind {
		t.Fatal("Me permission map changed")
	}
	m := userv1.File_user_v1_user_service_proto.Services().Get(0).Methods().ByName("Me")
	if m == nil || m.Input().FullName() != "google.protobuf.Empty" || m.Output().FullName() != "user.v1.UserData" || m.IsStreamingClient() || m.IsStreamingServer() {
		t.Fatal("Me RPC signature changed")
	}
}

func meAccessWire(t *testing.T) {
	for _, permissions := range []map[string]bool{nil, {}, {"iam.user.read": true, "clm.contract.read": true}} {
		input := &userv1.UserData{Id: 42, CountryCode: "+62", Phone: "+6281234567890", Status: "ACTIVE", Version: 7, Permissions: permissions}
		roleWire(t, input, new(userv1.UserData))
	}
}

func meAccessLegacy(t *testing.T) {
	file, err := protodesc.NewFile(beforeMePermissions(protodesc.ToFileDescriptorProto(userv1.File_user_v1_user_payload_proto)), protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}
	d := file.Messages().ByName("UserData")
	if d.Fields().Len() != 14 || d.Fields().ByNumber(15) != nil {
		t.Fatal("legacy Me descriptor changed")
	}
	input := &userv1.UserData{Id: 42, Email: "member@example.test", CountryCode: "+62", RoleCodes: []int32{1}, Permissions: map[string]bool{"clm.contract.read": true}}
	wire, err := proto.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	legacy := dynamicpb.NewMessage(d)
	if err = proto.Unmarshal(wire, legacy); err != nil || len(legacy.GetUnknown()) == 0 || legacy.Get(d.Fields().ByNumber(1)).Uint() != 42 {
		t.Fatal("legacy Me read lost fields or unknown map", err)
	}
	wire, err = proto.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	output := new(userv1.UserData)
	if err = proto.Unmarshal(wire, output); err != nil || !proto.Equal(input, output) {
		t.Fatal("legacy Me wire hop changed response", err)
	}
}

func meAccessRedaction(t *testing.T) {
	for _, input := range []string{`{"password":"private"}`, `{"passwordHash":"private"}`, `{"token":"private"}`, `{"refreshToken":"private"}`, `{"sessionId":"private"}`, `{"authVersion":1}`} {
		if err := protojson.Unmarshal([]byte(input), new(userv1.UserData)); err == nil {
			t.Fatal("unsafe Me field accepted", input)
		}
	}
}

// Strip only the independently tested additive map from frozen descriptor baselines.
func beforeMePermissions(file *descriptorpb.FileDescriptorProto) *descriptorpb.FileDescriptorProto {
	out := proto.Clone(file).(*descriptorpb.FileDescriptorProto)
	if out.GetName() != "user/v1/user.payload.proto" {
		return out
	}
	for _, message := range out.MessageType {
		if message.GetName() != "UserData" {
			continue
		}
		fields := message.Field[:0]
		for _, field := range message.Field {
			if field.GetName() != "permissions" {
				fields = append(fields, field)
			}
		}
		message.Field = fields
		nested := message.NestedType[:0]
		for _, entry := range message.NestedType {
			if entry.GetName() != "PermissionsEntry" {
				nested = append(nested, entry)
			}
		}
		message.NestedType = nested
	}
	return out
}
