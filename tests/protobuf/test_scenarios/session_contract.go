package test_scenarios

import (
	"bytes"
	paginatorv1 "github.com/siti-nabila/api-contracts/pb/paginator/v1"
	profilev1 "github.com/siti-nabila/api-contracts/pb/profile/v1"
	userv1 "github.com/siti-nabila/api-contracts/pb/user/v1"
	"github.com/siti-nabila/api-contracts/tests/protobuf/fixtures"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"testing"
)

func All() []testutils.Scenario {
	return []testutils.Scenario{
		{Name: "all legacy fields and RPC descriptors stay compatible", Run: legacyDescriptors},
		{Name: "legacy decoder preserves additive session response fields", Run: legacyTokenWire},
		{Name: "profile mask presence preserves legacy replacement wire", Run: profilePresence},
		{Name: "safe identity profile and list outputs contain no credential fields", Run: safeOutputs},
	}
}

func legacyFile(t *testing.T, name string) *descriptorpb.FileDescriptorProto {
	t.Helper()
	data, err := fixtures.Legacy.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	out := new(descriptorpb.FileDescriptorProto)
	if err = protojson.Unmarshal(data, out); err != nil {
		t.Fatal(err)
	}
	return out
}

func legacyDescriptors(t *testing.T) {
	files := map[string]protoreflect.FileDescriptor{
		"user.payload.json":      userv1.File_user_v1_user_payload_proto,
		"user.service.json":      userv1.File_user_v1_user_service_proto,
		"profile.payload.json":   profilev1.File_profile_v1_profile_payload_proto,
		"profile.service.json":   profilev1.File_profile_v1_profile_service_proto,
		"paginator.payload.json": paginatorv1.File_paginator_v1_paginator_payload_proto,
	}
	for name, current := range files {
		t.Run(name, func(t *testing.T) {
			old := legacyFile(t, name)
			now := protodesc.ToFileDescriptorProto(current)
			if old.GetName() != now.GetName() || old.GetPackage() != now.GetPackage() || old.GetSyntax() != now.GetSyntax() || !proto.Equal(old.GetOptions(), now.GetOptions()) {
				t.Fatal("file identity/options changed")
			}
			for _, msg := range old.GetMessageType() {
				found := current.Messages().ByName(protoreflect.Name(msg.GetName()))
				if found == nil {
					t.Fatalf("legacy message removed: %s", msg.GetName())
				}
				for _, field := range msg.GetField() {
					got := found.Fields().ByNumber(protoreflect.FieldNumber(field.GetNumber()))
					if got == nil || !proto.Equal(field, protodesc.ToFieldDescriptorProto(got)) {
						t.Fatalf("legacy %s field %d descriptor changed", msg.GetName(), field.GetNumber())
					}
				}
			}
			for _, enum := range old.GetEnumType() {
				found := current.Enums().ByName(protoreflect.Name(enum.GetName()))
				if found == nil || !proto.Equal(enum, protodesc.ToEnumDescriptorProto(found)) {
					t.Fatalf("legacy enum changed: %s", enum.GetName())
				}
			}
			if len(old.GetService()) != current.Services().Len() {
				t.Fatal("RPC service inventory changed")
			}
			for _, svc := range old.GetService() {
				got := current.Services().ByName(protoreflect.Name(svc.GetName()))
				if got == nil || !proto.Equal(svc, protodesc.ToServiceDescriptorProto(got)) {
					t.Fatalf("RPC signature/inventory changed: %s", svc.GetName())
				}
			}
		})
	}
}

func legacyTokenWire(t *testing.T) {
	want := &userv1.UserTokenResponse{
		Token: "access-fixture", RefreshToken: "refresh-fixture", SessionId: "00000000-0000-4000-8000-000000000001",
		AccessExpiresAt: &timestamppb.Timestamp{Seconds: 900}, SessionExpiresAt: &timestamppb.Timestamp{Seconds: 604800}, MustChangePassword: true,
	}
	wire, err := proto.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	old, err := protodesc.NewFile(legacyFile(t, "user.payload.json"), protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}
	legacy := dynamicpb.NewMessage(old.Messages().ByName("UserTokenResponse"))
	if err = proto.Unmarshal(wire, legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.Get(legacy.Descriptor().Fields().ByNumber(1)).String() != want.Token {
		t.Fatal("old token field lost")
	}
	if len(legacy.GetUnknown()) == 0 {
		t.Fatal("additive fields not retained by legacy decoder")
	}
	reencoded, err := proto.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	got := new(userv1.UserTokenResponse)
	if err = proto.Unmarshal(reencoded, got); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(want, got) {
		t.Fatal("additive credentials/session fields did not survive legacy wire hop")
	}
	register := &userv1.UserTokenResponse{RegistrationId: 42, RegistrationStatus: "PENDING"}
	wire, err = proto.Marshal(register)
	if err != nil {
		t.Fatal(err)
	}
	if err = proto.Unmarshal(wire, legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.Get(legacy.Descriptor().Fields().ByNumber(1)).String() != "" {
		t.Fatal("pending registration unexpectedly contains legacy token")
	}
	if register.GetRefreshToken() != "" || register.GetSessionId() != "" || register.GetAccessExpiresAt() != nil || register.GetSessionExpiresAt() != nil {
		t.Fatal("pending registration must not carry credential/session fields")
	}
}

func profilePresence(t *testing.T) {
	legacyWire := []byte{10, 3, 'A', 'n', 'a', 18, 1, 'x', 26, 1, 'y'}
	old := new(profilev1.ProfileRequest)
	if err := proto.Unmarshal(legacyWire, old); err != nil {
		t.Fatal(err)
	}
	if old.GetName() != "Ana" || old.GetAddress() != "x" || old.GetPhone() != "y" || old.GetFieldMask() != nil {
		t.Fatal("legacy replacement changed")
	}
	got, err := proto.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, legacyWire) {
		t.Fatal("legacy replacement wire changed")
	}
	for _, mask := range []*fieldmaskpb.FieldMask{{}, {Paths: []string{"address"}}} {
		input := &profilev1.ProfileRequest{FieldMask: mask}
		data, err := proto.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		output := new(profilev1.ProfileRequest)
		if err = proto.Unmarshal(data, output); err != nil {
			t.Fatal(err)
		}
		if output.FieldMask == nil || !proto.Equal(output.FieldMask, mask) || output.Address != "" {
			t.Fatal("present mask and selected-empty presence lost")
		}
	}
}

func safeOutputs(t *testing.T) {
	seen := map[protoreflect.FullName]bool{}
	var check func(protoreflect.MessageDescriptor)
	check = func(desc protoreflect.MessageDescriptor) {
		if seen[desc.FullName()] {
			return
		}
		seen[desc.FullName()] = true
		for i := 0; i < desc.Fields().Len(); i++ {
			field := desc.Fields().Get(i)
			name := string(field.Name())
			for _, unsafe := range []string{"password", "token", "secret", "hash", "auth_version"} {
				if strings.Contains(name, unsafe) && name != "must_change_password" {
					t.Errorf("credential field in safe output: %s", field.FullName())
				}
			}
			if field.Message() != nil {
				check(field.Message())
			}
		}
	}
	for _, msg := range []proto.Message{&userv1.UserData{}, &userv1.UserListItem{}, &profilev1.ProfileResponse{}} {
		check(msg.ProtoReflect().Descriptor())
	}
	// New identity metadata survives serialization without changing numeric role codes.
	want := &userv1.UserData{Id: 1, Status: "ACTIVE", Version: 3, RoleCodes: []int32{1}, Roles: []*userv1.UserRoleAssignment{{RoleId: 1, Code: "iam.member", Status: "ACTIVE", RoleStatus: "ACTIVE"}}, ApplicationAccess: []*userv1.UserApplicationAccess{{ApplicationId: 2, Code: "clm", Status: "ACTIVE", ApplicationStatus: "ACTIVE"}}}
	data, err := proto.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	got := new(userv1.UserData)
	if err = proto.Unmarshal(data, got); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(want, got) {
		t.Fatal("safe role/application metadata lost")
	}
}
