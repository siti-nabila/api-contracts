package test_scenarios

import (
	"encoding/json"
	applicationv1 "github.com/siti-nabila/api-contracts/pb/application/v1"
	paginatorv1 "github.com/siti-nabila/api-contracts/pb/paginator/v1"
	userv1 "github.com/siti-nabila/api-contracts/pb/user/v1"
	"github.com/siti-nabila/api-contracts/tests/protobuf/fixtures"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"strings"
	"testing"
)

func ApplicationAccess() []testutils.Scenario {
	return []testutils.Scenario{
		{Name: "all thirty baseline methods and complete messages remain compatible", Run: applicationAccessBaseline},
		{Name: "three exact unary methods and every new tag type and presence", Run: applicationAccessFields},
		{Name: "optional access version provenance and compound results survive binary and JSON", Run: applicationAccessWire},
		{Name: "query filters and all eleven public scalars are documented safely", Run: applicationAccessDocs},
	}
}

func applicationAccessBaseline(t *testing.T) {
	var files map[string]json.RawMessage
	if err := json.Unmarshal(fixtures.ApplicationAccessBaseline, &files); err != nil {
		t.Fatal(err)
	}
	oldCount, count := 0, 0
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
			count += len(service.Method)
		}
	}
	if len(files) != 16 || oldCount != 30 || count != 35 {
		t.Fatal("inventory", len(files), oldCount, count)
	}
}

func applicationAccessFields(t *testing.T) {
	for _, c := range []struct {
		file    protoreflect.FileDescriptor
		prefix  string
		methods map[string]string
		fields  map[string]string
	}{
		{userv1.File_user_v1_user_payload_proto, "user.v1.", map[string]string{"ListUserApplicationAccess": userv1.UserService_ListUserApplicationAccess_FullMethodName}, map[string]string{
			"UserApplicationAccessItem":         "application_id:uint64 application_code:string name:string application_status:string is_system:bool status:string granted_by:uint64? granted_at:google.protobuf.Timestamp revoked_by:uint64? revoked_at:google.protobuf.Timestamp revoke_reason:string?",
			"UserApplicationAccessFilter":       "application_code:string statuses:string*",
			"ListUserApplicationAccessRequest":  "user_id:uint64 query:paginator.v1.PageQuery filter:user.v1.UserApplicationAccessFilter",
			"ListUserApplicationAccessResponse": "user_id:uint64 version:uint64 items:user.v1.UserApplicationAccessItem* meta:paginator.v1.PageMeta next_cursor:string",
		}},
		{applicationv1.File_application_v1_application_payload_proto, "application.v1.", map[string]string{"GrantApplicationAccess": applicationv1.ApplicationService_GrantApplicationAccess_FullMethodName, "RevokeApplicationAccess": applicationv1.ApplicationService_RevokeApplicationAccess_FullMethodName}, map[string]string{
			"GrantApplicationAccessRequest":   "user_id:uint64 application_code:string role_ids:uint64* reason:string expected_user_version:uint64?",
			"RevokeApplicationAccessRequest":  "user_id:uint64 application_code:string reason:string expected_user_version:uint64?",
			"GrantApplicationAccessResponse":  "user_id:uint64 application_id:uint64 access_status:string? changed_role_ids:uint64* version:uint64 changed:bool",
			"RevokeApplicationAccessResponse": "user_id:uint64 application_id:uint64 access_status:string? changed_role_ids:uint64* version:uint64 changed:bool",
		}},
	} {
		serviceFile := userv1.File_user_v1_user_service_proto
		service := "UserService"
		if c.prefix == "application.v1." {
			serviceFile, service = applicationv1.File_application_v1_application_service_proto, "ApplicationService"
		}
		for name, full := range c.methods {
			method := serviceFile.Services().ByName(protoreflect.Name(service)).Methods().ByName(protoreflect.Name(name))
			if method == nil || method.IsStreamingClient() || method.IsStreamingServer() || string(method.Input().FullName()) != c.prefix+name+"Request" || string(method.Output().FullName()) != c.prefix+name+"Response" || full != "/"+c.prefix+service+"/"+name {
				t.Fatal(name)
			}
		}
		for name, fields := range c.fields {
			message := c.file.Messages().ByName(protoreflect.Name(name))
			want := strings.Fields(fields)
			if message == nil || message.Fields().Len() != len(want) {
				t.Fatal(name)
			}
			for i, field := range want {
				assertRoleField(t, message, i, field)
			}
		}
	}
}

func applicationAccessWire(t *testing.T) {
	for _, version := range []*uint64{nil, proto.Uint64(0), proto.Uint64(9223372036854775807)} {
		roleWire(t, &applicationv1.GrantApplicationAccessRequest{UserId: 7, ApplicationCode: "sales", RoleIds: []uint64{3, 2, 3}, Reason: "Review scope", ExpectedUserVersion: version}, new(applicationv1.GrantApplicationAccessRequest))
		roleWire(t, &applicationv1.RevokeApplicationAccessRequest{UserId: 7, ApplicationCode: "sales", Reason: "Review scope", ExpectedUserVersion: version}, new(applicationv1.RevokeApplicationAccessRequest))
	}
	for _, access := range []*string{nil, proto.String(""), proto.String("REVOKED")} {
		for _, id := range []*uint64{nil, proto.Uint64(0), proto.Uint64(4)} {
			roleWire(t, &userv1.UserApplicationAccessItem{ApplicationId: 2, Status: "REVOKED", GrantedBy: id, GrantedAt: &timestamppb.Timestamp{Seconds: 1735689600}, RevokedBy: id, RevokeReason: access}, new(userv1.UserApplicationAccessItem))
		}
		roleWire(t, &applicationv1.RevokeApplicationAccessResponse{UserId: 7, ApplicationId: 2, AccessStatus: access, Version: 9}, new(applicationv1.RevokeApplicationAccessResponse))
	}
	ids := make([]uint64, 25)
	for i := range ids {
		ids[i] = uint64(i + 1)
	}
	for _, input := range []proto.Message{
		&applicationv1.GrantApplicationAccessResponse{UserId: 7, ApplicationId: 2, AccessStatus: proto.String("ACTIVE"), Version: 9, Changed: true},
		&applicationv1.RevokeApplicationAccessResponse{UserId: 7, ApplicationId: 2, ChangedRoleIds: ids, Version: 9, Changed: true},
		&userv1.ListUserApplicationAccessResponse{UserId: 7, Version: 9, Items: []*userv1.UserApplicationAccessItem{{ApplicationId: 2}}, Meta: &paginatorv1.PageMeta{Page: 1, Limit: 20, Total: 1, TotalPages: 1}},
		&userv1.ListUserApplicationAccessRequest{UserId: 7, Filter: &userv1.UserApplicationAccessFilter{ApplicationCode: "sales", Statuses: []string{"ACTIVE", "REVOKED"}}, Query: &paginatorv1.PageQuery{Fields: []string{"name"}, Sort: []*paginatorv1.Sort{{Field: "granted_by", Desc: true}, {Field: "name"}}, Search: &paginatorv1.Search{Fields: []string{"name"}, Keyword: "50%_", Mode: paginatorv1.SearchMode_SEARCH_MODE_PREFIX}, LastId: "typed-cursor"}},
	} {
		roleWire(t, input, input.ProtoReflect().New().Interface())
	}
	for _, data := range []string{`{"actorId":1}`, `{"password":"private"}`, `{"authVersion":1}`, `{"roleIds":[1]}`} {
		if err := protojson.Unmarshal([]byte(data), new(applicationv1.RevokeApplicationAccessRequest)); err == nil {
			t.Fatal("undeclared revoke input", data)
		}
	}
}

func applicationAccessDocs(t *testing.T) {
	data, err := os.ReadFile("../../docs/application-access-contract.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	message := (&userv1.UserApplicationAccessItem{}).ProtoReflect().Descriptor()
	for i := 0; i < message.Fields().Len(); i++ {
		if strings.Count(text, "| "+string(message.Fields().Get(i).Name())+" |") != 1 {
			t.Fatal("public scalar missing/duplicate", message.Fields().Get(i).Name())
		}
	}
	for _, method := range []string{userv1.UserService_ListUserApplicationAccess_FullMethodName, applicationv1.ApplicationService_GrantApplicationAccess_FullMethodName, applicationv1.ApplicationService_RevokeApplicationAccess_FullMethodName} {
		if strings.Count(text, "| "+method+" |") != 1 {
			t.Fatal("method documentation", method)
		}
	}
}

// Baseline B assertions stay exact while allowing the separately frozen D additions.
func applicationAdditionalMethods(service string) int {
	if service == "ApplicationService" {
		return 2
	}
	return 0
}
func applicationAdditionalMessages(path string) int {
	if path == "application/v1/application.payload.proto" {
		return 4
	}
	return 0
}
