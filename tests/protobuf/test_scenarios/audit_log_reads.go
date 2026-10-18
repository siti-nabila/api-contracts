package test_scenarios

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	auditv1 "github.com/siti-nabila/api-contracts/pb/audit/v1"
	paginatorv1 "github.com/siti-nabila/api-contracts/pb/paginator/v1"
	"github.com/siti-nabila/api-contracts/tests/protobuf/fixtures"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func AuditLogReads() []testutils.Scenario {
	return []testutils.Scenario{
		{Name: "prior18 descriptors and36 methods remain unchanged with exact37 method inventory", Run: auditBaseline},
		{Name: "audit unary messages tags types and presence are fixed", Run: auditFields},
		{Name: "BIGINT nullable actors and typed snapshots roundtrip without unsafe fields", Run: auditWire},
		{Name: "audit minimal query and safe contract are documented", Run: auditDocs},
	}
}

func auditBaseline(t *testing.T) {
	var files map[string]json.RawMessage
	if err := json.Unmarshal(fixtures.AuditLogReadsBaseline, &files); err != nil {
		t.Fatal(err)
	}
	methods := 0
	for path, b := range files {
		old := new(descriptorpb.FileDescriptorProto)
		if err := protojson.Unmarshal(b, old); err != nil {
			t.Fatal(err)
		}
		f, err := protoregistry.GlobalFiles.FindFileByPath(path)
		if err != nil {
			t.Fatal(err)
		}
		if !proto.Equal(old, protodesc.ToFileDescriptorProto(f)) {
			t.Error("existing descriptor changed", path)
		}
		for _, s := range old.Service {
			methods += len(s.Method)
		}
	}
	if len(files) != 18 || methods != 36 {
		t.Fatal("baseline", len(files), methods)
	}
	services, total := 0, 0
	protoregistry.GlobalFiles.RangeFiles(func(f protoreflect.FileDescriptor) bool {
		services += f.Services().Len()
		for i := 0; i < f.Services().Len(); i++ {
			total += f.Services().Get(i).Methods().Len()
		}
		return true
	})
	if services != 9 || total != 37 {
		t.Fatal("inventory", services, total)
	}
}

func auditFields(t *testing.T) {
	s := auditv1.File_audit_v1_audit_service_proto.Services().Get(0)
	m := s.Methods().Get(0)
	if s.FullName() != "audit.v1.AuditService" || s.Methods().Len() != 1 || m.Name() != "ListAuditLogs" || m.IsStreamingClient() || m.IsStreamingServer() || m.Input().FullName() != "audit.v1.ListAuditLogsRequest" || m.Output().FullName() != "audit.v1.ListAuditLogsResponse" || auditv1.AuditService_ListAuditLogs_FullMethodName != "/audit.v1.AuditService/ListAuditLogs" {
		t.Fatal("unary contract", s.FullName(), m.FullName())
	}
	for name, fields := range map[string]string{
		"AuditLogFilter":        "actor_user_id:uint64? actor_service:string action:string target_type:string target_id:string occurred_from:google.protobuf.Timestamp occurred_before:google.protobuf.Timestamp",
		"AuditSnapshot":         "version:int64 status:string name:string? description:string? address:string? phone:string? role_ids:uint64* permission_codes:string* application_id:uint64 session_id:string reason_code:string no_op:bool",
		"AuditLogItem":          "id:uint64 actor_user_id:uint64? actor_service:string? action:string target_type:string target_id:string safe_before:audit.v1.AuditSnapshot safe_after:audit.v1.AuditSnapshot reason:string? request_id:string? trace_id:string? occurred_at:google.protobuf.Timestamp",
		"ListAuditLogsRequest":  "query:paginator.v1.PageQuery filter:audit.v1.AuditLogFilter",
		"ListAuditLogsResponse": "items:audit.v1.AuditLogItem* meta:paginator.v1.PageMeta next_cursor:string",
	} {
		d := auditv1.File_audit_v1_audit_payload_proto.Messages().ByName(protoreflect.Name(name))
		want := strings.Fields(fields)
		if d == nil || d.Fields().Len() != len(want) {
			t.Fatal("fields", name)
		}
		for i, v := range want {
			assertRoleField(t, d, i, v)
		}
	}
	for name, fields := range map[string][]string{"AuditLogFilter": {"actor_user_id"}, "AuditLogItem": {"actor_user_id", "actor_service", "reason", "request_id", "trace_id"}, "AuditSnapshot": {"name", "description", "address", "phone"}} {
		d := auditv1.File_audit_v1_audit_payload_proto.Messages().ByName(protoreflect.Name(name))
		for _, name := range fields {
			if !d.Fields().ByName(protoreflect.Name(name)).HasOptionalKeyword() {
				t.Error("optional presence", d.FullName(), name)
			}
		}
	}
}

func auditWire(t *testing.T) {
	at := timestamppb.New(time.Date(2026, 10, 18, 0, 0, 0, 123456000, time.UTC))
	for _, id := range []*uint64{nil, proto.Uint64(0), proto.Uint64(math.MaxInt32)} {
		roleWire(t, &auditv1.ListAuditLogsRequest{Query: &paginatorv1.PageQuery{Page: 2, Limit: 100, LastId: "opaque-cursor"}, Filter: &auditv1.AuditLogFilter{ActorUserId: id, ActorService: "service", Action: "USER_CREATED", TargetType: "user", TargetId: "7", OccurredFrom: at, OccurredBefore: at}}, new(auditv1.ListAuditLogsRequest))
	}
	for _, id := range []uint64{math.MaxInt32 + 1, math.MaxInt64} {
		for _, snapshot := range []*auditv1.AuditSnapshot{nil, {}, {Version: 3, Status: "ACTIVE", Name: proto.String(""), Description: proto.String("description"), Address: proto.String("address"), Phone: proto.String("000"), RoleIds: []uint64{2, 3}, PermissionCodes: []string{"iam.audit.read"}, ApplicationId: 1, SessionId: "00000000-0000-4000-8000-000000000001", ReasonCode: "REQUEST_DENIED", NoOp: true}} {
			roleWire(t, &auditv1.ListAuditLogsResponse{Items: []*auditv1.AuditLogItem{{Id: id, ActorUserId: proto.Uint64(7), ActorService: proto.String("service"), Action: "NO_OP", TargetType: "user", TargetId: "7", SafeBefore: snapshot, SafeAfter: &auditv1.AuditSnapshot{}, Reason: proto.String(""), RequestId: proto.String("request"), TraceId: proto.String("trace"), OccurredAt: at}}, Meta: &paginatorv1.PageMeta{Page: 1, Limit: 20, Total: 1, TotalPages: 1}, NextCursor: "opaque"}, new(auditv1.ListAuditLogsResponse))
		}
	}
	roleWire(t, &auditv1.ListAuditLogsResponse{Items: []*auditv1.AuditLogItem{{Id: 7, ActorService: proto.String("service"), Action: "SECURITY_ATTEMPT_FAILED", TargetType: "rpc", TargetId: auditv1.AuditService_ListAuditLogs_FullMethodName, OccurredAt: at}}}, new(auditv1.ListAuditLogsResponse))
	for _, s := range []string{`{"password":"synthetic"}`, `{"accessToken":"synthetic"}`, `{"refreshToken":"synthetic"}`, `{"safeBefore":{"tokenHash":"synthetic"}}`} {
		if err := protojson.Unmarshal([]byte(s), new(auditv1.AuditLogItem)); err == nil {
			t.Fatal("unsafe undeclared field", s)
		}
	}
}

func auditDocs(t *testing.T) {
	b, err := os.ReadFile("../../docs/audit-log-contract.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	for _, s := range []string{auditv1.AuditService_ListAuditLogs_FullMethodName, "iam.audit.read", "occurred_at DESC, id DESC", "INVALID_ARGUMENT", "UNAVAILABLE", "UNAUTHENTICATED", "PERMISSION_DENIED", "120/minute", "4096", "9223372036854775807", "no search", "no field selection"} {
		if !strings.Contains(text, s) {
			t.Error("missing documentation", s)
		}
	}
	for _, msg := range []proto.Message{new(auditv1.AuditLogFilter), new(auditv1.AuditSnapshot), new(auditv1.AuditLogItem), new(auditv1.ListAuditLogsRequest), new(auditv1.ListAuditLogsResponse)} {
		d := msg.ProtoReflect().Descriptor()
		for i := 0; i < d.Fields().Len(); i++ {
			if strings.Count(text, "| "+string(d.Name())+" | "+string(d.Fields().Get(i).Name())+" |") != 1 {
				t.Error("field docs", d.Fields().Get(i).FullName())
			}
		}
	}
}
