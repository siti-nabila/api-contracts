package test_scenarios

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	applicationv1 "github.com/siti-nabila/api-contracts/pb/application/v1"
	paginatorv1 "github.com/siti-nabila/api-contracts/pb/paginator/v1"
	permissionv1 "github.com/siti-nabila/api-contracts/pb/permission/v1"
	rolev1 "github.com/siti-nabila/api-contracts/pb/role/v1"
	"github.com/siti-nabila/api-contracts/tests/protobuf/fixtures"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func RoleManagement() []testutils.Scenario {
	return []testutils.Scenario{
		{Name: "all nineteen baseline methods and descriptors unchanged", Run: roleBaseline},
		{Name: "eight additive methods have exact unary signatures", Run: roleMethods},
		{Name: "all new public field tags types and presence frozen", Run: roleFields},
		{Name: "permission intent and mutation presence survive binary and JSON", Run: rolePresence},
		{Name: "safe response metadata and pagination survive the wire", Run: roleResponses},
		{Name: "documented scalar registry and methods match descriptors", Run: roleDocumentation},
	}
}

func roleBaseline(t *testing.T) {
	var files map[string]json.RawMessage
	if err := json.Unmarshal(fixtures.RoleBaseline, &files); err != nil {
		t.Fatal(err)
	}
	count := 0
	for path, data := range files {
		old := new(descriptorpb.FileDescriptorProto)
		if err := protojson.Unmarshal(data, old); err != nil {
			t.Fatal(err)
		}
		current, err := protoregistry.GlobalFiles.FindFileByPath(path)
		if err != nil {
			t.Fatal(err)
		}
		if !assignmentBaselineContains(old, protodesc.ToFileDescriptorProto(current)) {
			t.Error("baseline descriptor changed", path)
		}
		for _, s := range old.GetService() {
			count += len(s.GetMethod())
		}
	}
	if len(files) != 9 || count != 19 {
		t.Fatal("incomplete baseline inventory", len(files), count)
	}
}

func roleMethods(t *testing.T) {
	cases := []struct {
		file    protoreflect.FileDescriptor
		service string
		methods map[string]string
	}{
		{rolev1.File_role_v1_role_service_proto, "RoleService", map[string]string{
			"ListRoles": rolev1.RoleService_ListRoles_FullMethodName, "GetRole": rolev1.RoleService_GetRole_FullMethodName,
			"CreateRole": rolev1.RoleService_CreateRole_FullMethodName, "UpdateRole": rolev1.RoleService_UpdateRole_FullMethodName,
			"ChangeRoleStatus": rolev1.RoleService_ChangeRoleStatus_FullMethodName, "ReplaceRolePermissions": rolev1.RoleService_ReplaceRolePermissions_FullMethodName}},
		{permissionv1.File_permission_v1_permission_service_proto, "PermissionService", map[string]string{"ListPermissions": permissionv1.PermissionService_ListPermissions_FullMethodName}},
		{applicationv1.File_application_v1_application_service_proto, "ApplicationService", map[string]string{"ListApplications": applicationv1.ApplicationService_ListApplications_FullMethodName}},
	}
	for _, c := range cases {
		s := c.file.Services().ByName(protoreflect.Name(c.service))
		if c.file.Services().Len() != 1 || s == nil || s.Methods().Len() != len(c.methods)+applicationAdditionalMethods(c.service) {
			t.Fatal("service inventory", c.service)
		}
		for name, full := range c.methods {
			m := s.Methods().ByName(protoreflect.Name(name))
			prefix := string(c.file.Package()) + "."
			if m == nil || m.IsStreamingClient() || m.IsStreamingServer() || string(m.Input().FullName()) != prefix+name+"Request" || string(m.Output().FullName()) != prefix+name+"Response" || full != "/"+prefix+c.service+"/"+name {
				t.Error("method mismatch", name)
			}
		}
	}
}

func roleFields(t *testing.T) {
	cases := []struct {
		file   protoreflect.FileDescriptor
		fields map[string]string
	}{
		{rolev1.File_role_v1_role_payload_proto, map[string]string{
			"RoleSummary":                    "id:uint64 code:string name:string description:string? application_id:uint64 application_code:string status:string is_system:bool version:uint64 role_code:int32 created_at:google.protobuf.Timestamp updated_at:google.protobuf.Timestamp affected_user_count:uint64",
			"RoleFilter":                     "application_code:string statuses:string*",
			"ListRolesRequest":               "query:paginator.v1.PageQuery filter:role.v1.RoleFilter",
			"ListRolesResponse":              "items:role.v1.RoleSummary* meta:paginator.v1.PageMeta next_cursor:string",
			"GetRoleRequest":                 "role_id:uint64",
			"GetRoleResponse":                "role:role.v1.RoleSummary permissions:permission.v1.PermissionItem*",
			"CreateRoleRequest":              "application_code:string code:string name:string description:string? permission_codes:string*",
			"CreateRoleResponse":             "role:role.v1.RoleSummary permissions:permission.v1.PermissionItem*",
			"UpdateRoleRequest":              "role_id:uint64 expected_role_version:uint64? name:string? description:string? field_mask:google.protobuf.FieldMask",
			"UpdateRoleResponse":             "role:role.v1.RoleSummary changed:bool",
			"ChangeRoleStatusRequest":        "role_id:uint64 status:string? reason:string expected_role_version:uint64?",
			"ChangeRoleStatusResponse":       "role_id:uint64 status:string version:uint64 changed:bool",
			"PermissionSet":                  "codes:string*",
			"ReplaceRolePermissionsRequest":  "role_id:uint64 permissions:role.v1.PermissionSet reason:string expected_role_version:uint64?",
			"ReplaceRolePermissionsResponse": "role:role.v1.RoleSummary permissions:permission.v1.PermissionItem* changed:bool",
		}},
		{permissionv1.File_permission_v1_permission_payload_proto, map[string]string{
			"PermissionItem":          "id:uint64 application_id:uint64 application_code:string code:string description:string",
			"ListPermissionsRequest":  "query:paginator.v1.PageQuery application_code:string",
			"ListPermissionsResponse": "items:permission.v1.PermissionItem* meta:paginator.v1.PageMeta next_cursor:string",
		}},
		{applicationv1.File_application_v1_application_payload_proto, map[string]string{
			"ApplicationItem":          "id:uint64 code:string name:string status:string is_system:bool",
			"ApplicationFilter":        "statuses:string*",
			"ListApplicationsRequest":  "query:paginator.v1.PageQuery filter:application.v1.ApplicationFilter",
			"ListApplicationsResponse": "items:application.v1.ApplicationItem* meta:paginator.v1.PageMeta next_cursor:string",
		}},
	}
	for _, c := range cases {
		if c.file.Messages().Len() != len(c.fields)+applicationAdditionalMessages(c.file.Path()) {
			t.Fatal("new message inventory changed", c.file.Path())
		}
		for name, want := range c.fields {
			m := c.file.Messages().ByName(protoreflect.Name(name))
			fields := strings.Fields(want)
			if m == nil || m.Fields().Len() != len(fields) {
				t.Fatal("field inventory changed", name)
			}
			for i, spec := range fields {
				assertRoleField(t, m, i, spec)
			}
		}
	}
}

func assertRoleField(t *testing.T, m protoreflect.MessageDescriptor, i int, spec string) {
	t.Helper()
	parts := strings.SplitN(spec, ":", 2)
	typ := parts[1]
	optional := strings.HasSuffix(typ, "?")
	repeated := strings.HasSuffix(typ, "*")
	typ = strings.TrimRight(typ, "?*")
	f := m.Fields().ByNumber(protoreflect.FieldNumber(i + 1))
	kind := ""
	if f == nil {
		t.Fatal("missing tag", m.FullName(), i+1)
	}
	if f.Message() != nil {
		kind = string(f.Message().FullName())
	} else {
		kind = f.Kind().String()
	}
	if string(f.Name()) != parts[0] || kind != typ || f.HasOptionalKeyword() != optional || f.IsList() != repeated || f.HasPresence() != (optional || (f.Message() != nil && !repeated)) {
		t.Errorf("field contract mismatch %s: expected %s tag %d", f.FullName(), spec, i+1)
	}
}

func roleWire(t *testing.T, in, out proto.Message) {
	t.Helper()
	b, err := proto.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if err = proto.Unmarshal(b, out); err != nil || !proto.Equal(in, out) {
		t.Fatal("binary round trip", err)
	}
	b, err = protojson.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	proto.Reset(out)
	if err = protojson.Unmarshal(b, out); err != nil || !proto.Equal(in, out) {
		t.Fatal("JSON round trip", err)
	}
}

func rolePresence(t *testing.T) {
	absent := &rolev1.ReplaceRolePermissionsRequest{RoleId: 7}
	empty := &rolev1.ReplaceRolePermissionsRequest{RoleId: 7, Permissions: &rolev1.PermissionSet{}}
	a, _ := proto.Marshal(absent)
	b, _ := proto.Marshal(empty)
	if bytes.Equal(a, b) {
		t.Fatal("absent intent equals explicit empty intent")
	}
	for _, in := range []*rolev1.ReplaceRolePermissionsRequest{absent, empty, {RoleId: 7, Permissions: &rolev1.PermissionSet{Codes: []string{"clm.contract.read", "clm.contract.update"}}, ExpectedRoleVersion: proto.Uint64(3), Reason: "Review scope"}} {
		out := new(rolev1.ReplaceRolePermissionsRequest)
		roleWire(t, in, out)
		if (in.Permissions == nil) != (out.Permissions == nil) {
			t.Fatal("permission intent lost")
		}
	}
	for _, version := range []*uint64{nil, proto.Uint64(0), proto.Uint64(3)} {
		for _, text := range []*string{nil, proto.String(""), proto.String("Reviewer")} {
			roleWire(t, &rolev1.UpdateRoleRequest{RoleId: 7, ExpectedRoleVersion: version, Name: text, Description: proto.String(""), FieldMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description"}}}, new(rolev1.UpdateRoleRequest))
			roleWire(t, &rolev1.ChangeRoleStatusRequest{RoleId: 7, ExpectedRoleVersion: version, Status: text, Reason: "Policy review"}, new(rolev1.ChangeRoleStatusRequest))
		}
	}
	roleWire(t, &rolev1.CreateRoleRequest{ApplicationCode: "clm", Code: "clm.reviewer", Name: "Reviewer", Description: proto.String(""), PermissionCodes: []string{"clm.contract.read"}}, new(rolev1.CreateRoleRequest))
}

func roleResponses(t *testing.T) {
	r := &rolev1.RoleSummary{Id: 7, Code: "clm.reviewer", Name: "Reviewer", ApplicationId: 2, ApplicationCode: "clm", Status: "ACTIVE", Version: 3, RoleCode: 14, AffectedUserCount: 5}
	p := &permissionv1.PermissionItem{Id: 9, ApplicationId: 2, ApplicationCode: "clm", Code: "clm.contract.read", Description: "Read contracts"}
	meta := &paginatorv1.PageMeta{Page: 2, Limit: 20, Total: 51, TotalPages: 3, HasNext: true, HasPrev: true}
	cases := []proto.Message{
		&rolev1.ListRolesResponse{Items: []*rolev1.RoleSummary{r}, Meta: meta, NextCursor: "typed-cursor"},
		&rolev1.GetRoleResponse{Role: r, Permissions: []*permissionv1.PermissionItem{p}},
		&rolev1.CreateRoleResponse{Role: r, Permissions: []*permissionv1.PermissionItem{p}},
		&rolev1.UpdateRoleResponse{Role: r, Changed: true},
		&rolev1.ChangeRoleStatusResponse{RoleId: 7, Status: "DISABLED", Version: 4, Changed: true},
		&rolev1.ReplaceRolePermissionsResponse{Role: r, Permissions: []*permissionv1.PermissionItem{p}, Changed: true},
		&permissionv1.ListPermissionsResponse{Items: []*permissionv1.PermissionItem{p}, Meta: meta, NextCursor: "typed-cursor"},
		&applicationv1.ListApplicationsResponse{Items: []*applicationv1.ApplicationItem{{Id: 2, Code: "clm", Name: "Contracts", Status: "ACTIVE"}}, Meta: meta, NextCursor: "typed-cursor"},
	}
	for _, in := range cases {
		roleWire(t, in, in.ProtoReflect().New().Interface())
	}
	for _, d := range []*string{nil, proto.String(""), proto.String("Review contracts")} {
		r.Description = d
		for _, ts := range []*timestamppb.Timestamp{nil, {Seconds: 1735689600}} {
			r.CreatedAt = ts
			r.UpdatedAt = ts
			roleWire(t, r, new(rolev1.RoleSummary))
		}
	}
	// Projection may omit nullable fields while count=0 remains a valid scalar.
	roleWire(t, &rolev1.RoleSummary{Id: 7}, new(rolev1.RoleSummary))
}

func roleDocumentation(t *testing.T) {
	data, err := os.ReadFile("../../docs/role-management-contract.md")
	if err != nil {
		t.Fatal(err)
	}
	accessDoc, err := os.ReadFile("../../docs/application-access-contract.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data) + "\n" + string(accessDoc)
	for _, s := range []protoreflect.ServiceDescriptor{
		rolev1.File_role_v1_role_service_proto.Services().Get(0),
		permissionv1.File_permission_v1_permission_service_proto.Services().Get(0),
		applicationv1.File_application_v1_application_service_proto.Services().Get(0),
	} {
		for i := 0; i < s.Methods().Len(); i++ {
			method := "/" + string(s.FullName()) + "/" + string(s.Methods().Get(i).Name())
			if strings.Count(text, "| "+method+" |") != 1 {
				t.Error("method documentation missing/duplicate", method)
			}
		}
	}
	for _, m := range []protoreflect.MessageDescriptor{
		(&rolev1.RoleSummary{}).ProtoReflect().Descriptor(),
		(&permissionv1.PermissionItem{}).ProtoReflect().Descriptor(),
		(&applicationv1.ApplicationItem{}).ProtoReflect().Descriptor(),
	} {
		prefix := "| " + string(m.Name()) + " | "
		count := 0
		for _, line := range strings.Split(text, "\n") {
			if !strings.HasPrefix(line, prefix) {
				continue
			}
			count++
			cells := strings.Split(line, "|")
			if len(cells) != 6 {
				t.Fatal("malformed scalar documentation", line)
			}
			f := m.Fields().ByName(protoreflect.Name(strings.TrimSpace(cells[2])))
			if f == nil {
				t.Fatal("unknown documented scalar", line)
			}
			typ := f.Kind().String()
			if f.Message() != nil {
				typ = string(f.Message().FullName())
			}
			nullable := "No"
			if f.HasPresence() {
				nullable = "Yes"
			}
			if strings.TrimSpace(cells[3]) != typ || strings.TrimSpace(cells[4]) != nullable {
				t.Error("documented scalar type/null mismatch", line)
			}
		}
		if count != m.Fields().Len() {
			t.Error("documented scalar registry incomplete", m.FullName(), count)
		}
	}
	// Management contracts cannot accept undeclared caller authority or immutable fields.
	for _, in := range []string{
		`{"actorId":7}`, `{"isSystem":true}`, `{"roleCode":2}`, `{"status":"ACTIVE"}`,
	} {
		if err := protojson.Unmarshal([]byte(in), new(rolev1.CreateRoleRequest)); err == nil {
			t.Error("accepted undeclared creation input", in)
		}
	}
	for _, in := range []string{`{"code":"iam.admin"}`, `{"applicationCode":"iam"}`, `{"status":"ACTIVE"}`} {
		if err := protojson.Unmarshal([]byte(in), new(rolev1.UpdateRoleRequest)); err == nil {
			t.Error("accepted immutable update field", in)
		}
	}
}
