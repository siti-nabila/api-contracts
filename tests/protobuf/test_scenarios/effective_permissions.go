package test_scenarios

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	userv1 "github.com/siti-nabila/api-contracts/pb/user/v1"
	"github.com/siti-nabila/api-contracts/tests/protobuf/fixtures"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func EffectivePermissions() []testutils.Scenario {
	return []testutils.Scenario{
		{Name: "all33 baseline methods and complete messages remain compatible", Run: effectivePermissionsBaseline},
		{Name: "two exact unary signatures and all safe field tags types", Run: effectivePermissionsFields},
		{Name: "optional filters and inactive-account role sources survive binary and JSON", Run: effectivePermissionsWire},
		{Name: "contract documents identity status namespace and descriptive semantics", Run: effectivePermissionsDocs},
	}
}

func effectivePermissionsBaseline(t *testing.T) {
	var files map[string]json.RawMessage
	if err := json.Unmarshal(fixtures.EffectivePermissionsBaseline, &files); err != nil {
		t.Fatal(err)
	}
	oldCount, newCount := 0, 0
	for path, data := range files {
		old := new(descriptorpb.FileDescriptorProto)
		if err := protojson.Unmarshal(data, old); err != nil {
			t.Fatal(err)
		}
		file, err := protoregistry.GlobalFiles.FindFileByPath(path)
		if err != nil {
			t.Fatal(err)
		}
		current := protodesc.ToFileDescriptorProto(file)
		if !assignmentBaselineContains(old, current) {
			t.Error("baseline changed", path)
		}
		for _, service := range old.Service {
			oldCount += len(service.Method)
		}
		for _, service := range current.Service {
			newCount += len(service.Method)
		}
	}
	if len(files) != 16 || oldCount != 33 || newCount != 35 {
		t.Fatal("inventory", len(files), oldCount, newCount)
	}
}

func effectivePermissionsFields(t *testing.T) {
	service := userv1.File_user_v1_user_service_proto.Services().Get(0)
	for name, full := range map[string]string{
		"GetMyEffectivePermissions":   userv1.UserService_GetMyEffectivePermissions_FullMethodName,
		"GetUserEffectivePermissions": userv1.UserService_GetUserEffectivePermissions_FullMethodName,
	} {
		m := service.Methods().ByName(protoreflect.Name(name))
		if m == nil || m.IsStreamingClient() || m.IsStreamingServer() || string(m.Input().FullName()) != "user.v1."+name+"Request" || string(m.Output().FullName()) != "user.v1.GetEffectivePermissionsResponse" || full != "/user.v1.UserService/"+name {
			t.Fatal("method contract", name)
		}
	}
	for name, fields := range map[string]string{
		"GetMyEffectivePermissionsRequest":   "application_code:string",
		"GetUserEffectivePermissionsRequest": "user_id:uint64 application_code:string",
		"EffectivePermissionRole":            "role_id:uint64 code:string",
		"EffectivePermission":                "code:string roles:user.v1.EffectivePermissionRole*",
		"GetEffectivePermissionsResponse":    "user_id:uint64 user_status:string version:uint64 permissions:user.v1.EffectivePermission*",
	} {
		m := userv1.File_user_v1_user_payload_proto.Messages().ByName(protoreflect.Name(name))
		want := strings.Fields(fields)
		if m == nil || m.Fields().Len() != len(want) {
			t.Fatal("message fields", name)
		}
		for i, field := range want {
			assertRoleField(t, m, i, field)
		}
	}
}

func effectivePermissionsWire(t *testing.T) {
	for _, application := range []string{"", "iam", "sales", "unknown"} {
		roleWire(t, &userv1.GetMyEffectivePermissionsRequest{ApplicationCode: application}, new(userv1.GetMyEffectivePermissionsRequest))
		roleWire(t, &userv1.GetUserEffectivePermissionsRequest{UserId: 2147483647, ApplicationCode: application}, new(userv1.GetUserEffectivePermissionsRequest))
	}
	for _, status := range []string{"PENDING", "ACTIVE", "SUSPENDED", "DEACTIVATED"} {
		for _, permissions := range [][]*userv1.EffectivePermission{nil, {}, {{Code: "iam.user.read", Roles: []*userv1.EffectivePermissionRole{{RoleId: 2, Code: "iam.platform_admin"}, {RoleId: 3, Code: "iam.auditor"}}}}} {
			roleWire(t, &userv1.GetEffectivePermissionsResponse{UserId: 7, UserStatus: status, Version: 9, Permissions: permissions}, new(userv1.GetEffectivePermissionsResponse))
		}
	}
	for _, input := range []string{`{"userId":7}`, `{"actorId":7}`, `{"password":"private"}`, `{"roleCodes":[2]}`, `{"expectedUserVersion":1}`} {
		if err := protojson.Unmarshal([]byte(input), new(userv1.GetMyEffectivePermissionsRequest)); err == nil {
			t.Fatal("undeclared self input", input)
		}
	}
	for _, input := range []string{`{"token":"private"}`, `{"refreshToken":"private"}`, `{"password":"private"}`, `{"authVersion":1}`, `{"sessionId":"private"}`} {
		if err := protojson.Unmarshal([]byte(input), new(userv1.GetEffectivePermissionsResponse)); err == nil {
			t.Fatal("unsafe response field", input)
		}
	}
}

func effectivePermissionsDocs(t *testing.T) {
	data, err := os.ReadFile("../../docs/effective-permission-contract.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, method := range []string{userv1.UserService_GetMyEffectivePermissions_FullMethodName, userv1.UserService_GetUserEffectivePermissions_FullMethodName} {
		if strings.Count(text, "| "+method+" |") != 1 {
			t.Fatal("method documentation", method)
		}
	}
	for _, message := range []proto.Message{new(userv1.GetMyEffectivePermissionsRequest), new(userv1.GetUserEffectivePermissionsRequest), new(userv1.EffectivePermissionRole), new(userv1.EffectivePermission), new(userv1.GetEffectivePermissionsResponse)} {
		m := message.ProtoReflect().Descriptor()
		for i := 0; i < m.Fields().Len(); i++ {
			if strings.Count(text, "| "+string(m.Name())+" | "+string(m.Fields().Get(i).Name())+" |") != 1 {
				t.Fatal("field documentation", m.FullName(), m.Fields().Get(i).Name())
			}
		}
	}
}
