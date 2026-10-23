package test_scenarios

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	authorizationv1 "github.com/siti-nabila/api-contracts/pb/authorization/v1"
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

func AccessEvaluation() []testutils.Scenario {
	return []testutils.Scenario{
		{Name: "complete pre-evaluation descriptors and all35 methods unchanged", Run: evaluationBaseline},
		{Name: "exact unary signature fields tags types and enum numbers", Run: evaluationFields},
		{Name: "safe true and false decisions survive binary and JSON", Run: evaluationWire},
		{Name: "contract documents all fields trust boundary and bounded reasons", Run: evaluationDocs},
	}
}

func evaluationBaseline(t *testing.T) {
	var files map[string]json.RawMessage
	if err := json.Unmarshal(fixtures.AccessEvaluationBaseline, &files); err != nil {
		t.Fatal(err)
	}
	methods := 0
	for path, data := range files {
		old := new(descriptorpb.FileDescriptorProto)
		if err := protojson.Unmarshal(data, old); err != nil {
			t.Fatal(err)
		}
		current, err := protoregistry.GlobalFiles.FindFileByPath(path)
		if err != nil {
			t.Fatal(err)
		}
		if !proto.Equal(old, beforeCountryCode(protodesc.ToFileDescriptorProto(current))) {
			t.Error("existing descriptor changed", path)
		}
		for _, s := range old.Service {
			methods += len(s.Method)
		}
	}
	if len(files) != 16 || methods != 35 {
		t.Fatal("baseline", len(files), methods)
	}
	services, total := 0, 0
	protoregistry.GlobalFiles.RangeFiles(func(f protoreflect.FileDescriptor) bool {
		for i := 0; i < f.Services().Len(); i++ {
			services++
			total += f.Services().Get(i).Methods().Len()
		}
		return true
	})
	if services != 9 || total != 37 {
		t.Fatal("new inventory", services, total)
	}
}

func evaluationFields(t *testing.T) {
	f := authorizationv1.File_authorization_v1_authorization_service_proto
	s := f.Services().Get(0)
	m := s.Methods().Get(0)
	if s.FullName() != "authorization.v1.AuthorizationService" || s.Methods().Len() != 1 || m.Name() != "EvaluateAccess" || m.IsStreamingClient() || m.IsStreamingServer() || m.Input().FullName() != "authorization.v1.EvaluateAccessRequest" || m.Output().FullName() != "authorization.v1.EvaluateAccessResponse" || authorizationv1.AuthorizationService_EvaluateAccess_FullMethodName != "/authorization.v1.AuthorizationService/EvaluateAccess" {
		t.Fatal("unary contract", s.FullName(), m.FullName())
	}
	for name, fields := range map[string]string{
		"EvaluateAccessRequest":  "application_code:string permission_code:string",
		"EvaluateAccessResponse": "allowed:bool reason:enum user_id:uint64 sid:string user_version:uint64 evaluated_at:google.protobuf.Timestamp",
	} {
		m := authorizationv1.File_authorization_v1_authorization_payload_proto.Messages().ByName(protoreflect.Name(name))
		want := strings.Fields(fields)
		if m == nil || m.Fields().Len() != len(want) {
			t.Fatal("fields", name)
		}
		for i, v := range want {
			assertRoleField(t, m, i, v)
		}
	}
	e := authorizationv1.AccessDecisionReason(0).Descriptor()
	names := []string{"UNSPECIFIED", "ALLOWED", "PERMISSION_MISSING", "APPLICATION_INACTIVE", "APPLICATION_ACCESS_REQUIRED"}
	if e.Values().Len() != len(names) || m.Output().Fields().ByName("reason").Enum().FullName() != e.FullName() {
		t.Fatal("reason type")
	}
	for i, name := range names {
		v := e.Values().ByNumber(protoreflect.EnumNumber(i))
		if v == nil || string(v.Name()) != "ACCESS_DECISION_REASON_"+name {
			t.Fatal("enum contract", i, v)
		}
	}
}

func evaluationWire(t *testing.T) {
	roleWire(t, &authorizationv1.EvaluateAccessRequest{ApplicationCode: "iam", PermissionCode: "iam.user.status.update"}, new(authorizationv1.EvaluateAccessRequest))
	for reason := authorizationv1.AccessDecisionReason(1); reason <= 4; reason++ {
		roleWire(t, &authorizationv1.EvaluateAccessResponse{Allowed: reason == 1, Reason: reason, UserId: 2147483647, Sid: "00000000-0000-4000-8000-000000000001", UserVersion: 9, EvaluatedAt: timestamppb.New(time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC))}, new(authorizationv1.EvaluateAccessResponse))
	}
	for _, input := range []string{`{"userId":7}`, `{"sid":"private"}`, `{"roleCodes":[2]}`, `{"expectedUserVersion":1}`} {
		if err := protojson.Unmarshal([]byte(input), new(authorizationv1.EvaluateAccessRequest)); err == nil {
			t.Fatal("undeclared request field", input)
		}
	}
	for _, input := range []string{`{"token":"private"}`, `{"refreshToken":"private"}`, `{"password":"private"}`, `{"permissions":[]}`, `{"authVersion":1}`} {
		if err := protojson.Unmarshal([]byte(input), new(authorizationv1.EvaluateAccessResponse)); err == nil {
			t.Fatal("unsafe response field", input)
		}
	}
}

func evaluationDocs(t *testing.T) {
	b, err := os.ReadFile("../../docs/access-evaluation-contract.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	for _, required := range []string{authorizationv1.AuthorizationService_EvaluateAccess_FullMethodName, "PERMISSION_MISSING", "APPLICATION_INACTIVE", "APPLICATION_ACCESS_REQUIRED", "UNAVAILABLE", "UNAUTHENTICATED", "PERMISSION_DENIED", "120/minute", "not a policy revision", "unknown application"} {
		if !strings.Contains(text, required) {
			t.Error("missing contract documentation", required)
		}
	}
	for _, message := range []proto.Message{new(authorizationv1.EvaluateAccessRequest), new(authorizationv1.EvaluateAccessResponse)} {
		m := message.ProtoReflect().Descriptor()
		for i := 0; i < m.Fields().Len(); i++ {
			if strings.Count(text, "| "+string(m.Name())+" | "+string(m.Fields().Get(i).Name())+" |") != 1 {
				t.Error("field documentation", m.Fields().Get(i).FullName())
			}
		}
	}
}
