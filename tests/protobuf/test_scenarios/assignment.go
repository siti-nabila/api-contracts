package test_scenarios

import (
	"encoding/json"
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

func Assignment() []testutils.Scenario {
	return []testutils.Scenario{
		{Name: "27 baseline methods and every legacy field remain compatible", Run: assignmentBaseline},
		{Name: "three exact unary RPCs and all new tags types and presence", Run: assignmentFields},
		{Name: "version and nullable provenance survive binary and JSON", Run: assignmentWire},
		{Name: "all seventeen public scalars documented", Run: assignmentDocs},
	}
}

// Compare complete old messages and individual old methods, permitting additive
// top-level messages/methods only. No legacy field or option can disappear.
func assignmentBaselineContains(old, now *descriptorpb.FileDescriptorProto) bool {
	now = beforeCountryCode(now)
	if old.GetName() != now.GetName() || old.GetPackage() != now.GetPackage() || old.GetSyntax() != now.GetSyntax() || !proto.Equal(old.Options, now.Options) {
		return false
	}
	for _, m := range old.MessageType {
		found := false
		for _, n := range now.MessageType {
			if m.GetName() == n.GetName() {
				found = proto.Equal(m, n)
			}
		}
		if !found {
			return false
		}
	}
	for _, e := range old.EnumType {
		found := false
		for _, n := range now.EnumType {
			if e.GetName() == n.GetName() {
				found = proto.Equal(e, n)
			}
		}
		if !found {
			return false
		}
	}
	for _, s := range old.Service {
		var current *descriptorpb.ServiceDescriptorProto
		for _, n := range now.Service {
			if s.GetName() == n.GetName() {
				current = n
			}
		}
		if current == nil || !proto.Equal(s.Options, current.Options) {
			return false
		}
		for _, m := range s.Method {
			found := false
			for _, n := range current.Method {
				if m.GetName() == n.GetName() {
					found = proto.Equal(m, n)
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}
func assignmentBaseline(t *testing.T) {
	var files map[string]json.RawMessage
	if e := json.Unmarshal(fixtures.AssignmentBaseline, &files); e != nil {
		t.Fatal(e)
	}
	oldCount, currentCount := 0, 0
	for path, b := range files {
		old := new(descriptorpb.FileDescriptorProto)
		if e := protojson.Unmarshal(b, old); e != nil {
			t.Fatal(e)
		}
		f, e := protoregistry.GlobalFiles.FindFileByPath(path)
		if e != nil {
			t.Fatal(e)
		}
		n := protodesc.ToFileDescriptorProto(f)
		if !assignmentBaselineContains(old, n) {
			t.Error("incompatible baseline", path)
		}
		for _, s := range old.Service {
			oldCount += len(s.Method)
		}
		for _, s := range n.Service {
			currentCount += len(s.Method)
		}
	}
	if len(files) != 16 || oldCount != 27 || currentCount != 35 {
		t.Fatal("inventory", len(files), oldCount, currentCount)
	}
}
func assignmentFields(t *testing.T) {
	s := userv1.File_user_v1_user_service_proto.Services().Get(0)
	for name, full := range map[string]string{"ListUserRoles": userv1.UserService_ListUserRoles_FullMethodName, "GrantUserRoles": userv1.UserService_GrantUserRoles_FullMethodName, "RevokeUserRoles": userv1.UserService_RevokeUserRoles_FullMethodName} {
		m := s.Methods().ByName(protoreflect.Name(name))
		if m == nil || m.IsStreamingClient() || m.IsStreamingServer() || string(m.Input().FullName()) != "user.v1."+name+"Request" || string(m.Output().FullName()) != "user.v1."+name+"Response" || full != "/user.v1.UserService/"+name {
			t.Fatal(name)
		}
	}
	for name, want := range map[string]string{
		"UserRoleAssignmentItem":   "id:uint64 role_id:uint64 code:string role_code:int32 name:string status:string role_status:string application_id:uint64 application_code:string application_status:string is_system:bool application_access_status:string? granted_by:uint64? granted_at:google.protobuf.Timestamp revoked_by:uint64? revoked_at:google.protobuf.Timestamp revoke_reason:string?",
		"UserRoleAssignmentFilter": "application_code:string statuses:string*",
		"ListUserRolesRequest":     "user_id:uint64 query:paginator.v1.PageQuery filter:user.v1.UserRoleAssignmentFilter",
		"ListUserRolesResponse":    "items:user.v1.UserRoleAssignmentItem* meta:paginator.v1.PageMeta next_cursor:string user_id:uint64 version:uint64",
		"GrantUserRolesRequest":    "user_id:uint64 role_ids:uint64* reason:string expected_user_version:uint64?",
		"RevokeUserRolesRequest":   "user_id:uint64 role_ids:uint64* reason:string expected_user_version:uint64?",
		"GrantUserRolesResponse":   "user_id:uint64 changed_role_ids:uint64* version:uint64 changed:bool",
		"RevokeUserRolesResponse":  "user_id:uint64 changed_role_ids:uint64* version:uint64 changed:bool",
	} {
		m := userv1.File_user_v1_user_payload_proto.Messages().ByName(protoreflect.Name(name))
		fs := strings.Fields(want)
		if m == nil || m.Fields().Len() != len(fs) {
			t.Fatal(name)
		}
		for i, f := range fs {
			assertRoleField(t, m, i, f)
		}
	}
}
func assignmentWire(t *testing.T) {
	for _, v := range []*uint64{nil, proto.Uint64(0), proto.Uint64(3)} {
		roleWire(t, &userv1.GrantUserRolesRequest{UserId: 7, RoleIds: []uint64{2, 1, 2}, Reason: "Review access", ExpectedUserVersion: v}, new(userv1.GrantUserRolesRequest))
		roleWire(t, &userv1.RevokeUserRolesRequest{UserId: 7, RoleIds: []uint64{1}, Reason: "Review access", ExpectedUserVersion: v}, new(userv1.RevokeUserRolesRequest))
	}
	for _, text := range []*string{nil, proto.String(""), proto.String("REVOKED")} {
		for _, id := range []*uint64{nil, proto.Uint64(0), proto.Uint64(4)} {
			for _, ts := range []*timestamppb.Timestamp{nil, {Seconds: 1735689600}} {
				roleWire(t, &userv1.UserRoleAssignmentItem{Id: 1, RoleId: 2, ApplicationAccessStatus: text, GrantedBy: id, GrantedAt: ts, RevokedBy: id, RevokedAt: ts, RevokeReason: text}, new(userv1.UserRoleAssignmentItem))
			}
		}
	}
	for _, in := range []proto.Message{&userv1.GrantUserRolesResponse{UserId: 7, ChangedRoleIds: []uint64{1, 2}, Version: 4, Changed: true}, &userv1.RevokeUserRolesResponse{UserId: 7, Version: 4}, &userv1.ListUserRolesResponse{UserId: 7, Version: 4, Items: []*userv1.UserRoleAssignmentItem{{Id: 1, RoleId: 2}}}} {
		roleWire(t, in, in.ProtoReflect().New().Interface())
	}
}
func assignmentDocs(t *testing.T) {
	b, e := os.ReadFile("../../docs/user-role-assignment-contract.md")
	if e != nil {
		t.Fatal(e)
	}
	text := string(b)
	m := (&userv1.UserRoleAssignmentItem{}).ProtoReflect().Descriptor()
	for i := 0; i < m.Fields().Len(); i++ {
		if strings.Count(text, "| "+string(m.Fields().Get(i).Name())+" |") != 1 {
			t.Error("scalar documentation", m.Fields().Get(i).Name())
		}
	}
}
