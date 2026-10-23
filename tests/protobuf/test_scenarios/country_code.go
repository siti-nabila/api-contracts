package test_scenarios

import (
	"testing"

	auditv1 "github.com/siti-nabila/api-contracts/pb/audit/v1"
	profilev1 "github.com/siti-nabila/api-contracts/pb/profile/v1"
	userv1 "github.com/siti-nabila/api-contracts/pb/user/v1"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func CountryCode() []testutils.Scenario {
	return []testutils.Scenario{
		{Name: "calling code fields use new string tags and preserve input presence", Run: countryCodeFields},
		{Name: "omitted empty and populated calling codes survive protobuf roundtrips", Run: countryCodePresence},
		{Name: "calling code response fields survive legacy wire hops", Run: countryCodeCompatibility},
	}
}

func countryCodeFields(t *testing.T) {
	for _, c := range []struct {
		message  proto.Message
		tag      protoreflect.FieldNumber
		optional bool
	}{
		{new(profilev1.Profile), 6, false}, {new(profilev1.ProfileRequest), 5, true},
		{new(userv1.User), 14, false}, {new(userv1.UserData), 14, false},
		{new(userv1.UserListItem), 11, false}, {new(userv1.CreateUserRequest), 7, true},
		{new(userv1.UpdateUserRequest), 7, true}, {new(auditv1.AuditSnapshot), 13, true},
	} {
		d := c.message.ProtoReflect().Descriptor()
		f := d.Fields().ByName("country_code")
		if f == nil || f.Number() != c.tag || f.Kind() != protoreflect.StringKind || f.HasOptionalKeyword() != c.optional {
			t.Fatal("calling code descriptor changed", d.FullName())
		}
	}
}

func countryCodePresence(t *testing.T) {
	for _, value := range []*string{nil, proto.String(""), proto.String("+62")} {
		for _, input := range []proto.Message{
			&profilev1.ProfileRequest{CountryCode: value, FieldMask: &fieldmaskpb.FieldMask{Paths: []string{"country_code"}}},
			&userv1.CreateUserRequest{CountryCode: value},
			&userv1.UpdateUserRequest{CountryCode: value, FieldMask: &fieldmaskpb.FieldMask{Paths: []string{"country_code"}}},
			&auditv1.AuditSnapshot{CountryCode: value},
		} {
			output := input.ProtoReflect().New().Interface()
			managementRoundTrip(t, input, output)
			f := output.ProtoReflect().Descriptor().Fields().ByName("country_code")
			if output.ProtoReflect().Has(f) != (value != nil) {
				t.Fatal("calling code presence lost", f.FullName())
			}
		}
	}
}

func countryCodeCompatibility(t *testing.T) {
	for _, c := range []struct {
		file  string
		input proto.Message
	}{
		{"profile.payload.json", &profilev1.Profile{Id: 7, UserId: 42, CountryCode: "+62"}},
		{"profile.payload.json", &profilev1.ProfileRequest{Name: "Ana", CountryCode: proto.String("+62")}},
		{"user.payload.json", &userv1.User{Id: 42, CountryCode: "+62"}},
		{"user.payload.json", &userv1.UserData{Id: 42, CountryCode: "+62"}},
		{"user.payload.json", &userv1.UserListItem{Id: 42, CountryCode: "+62"}},
	} {
		old, err := protodesc.NewFile(legacyFile(t, c.file), protoregistry.GlobalFiles)
		if err != nil {
			t.Fatal(err)
		}
		d := old.Messages().ByName(c.input.ProtoReflect().Descriptor().Name())
		legacy := dynamicpb.NewMessage(d)
		wire, err := proto.Marshal(c.input)
		if err != nil {
			t.Fatal(err)
		}
		if err = proto.Unmarshal(wire, legacy); err != nil || len(legacy.GetUnknown()) == 0 {
			t.Fatal("additive calling code lost", err)
		}
		wire, err = proto.Marshal(legacy)
		if err != nil {
			t.Fatal(err)
		}
		output := c.input.ProtoReflect().New().Interface()
		if err = proto.Unmarshal(wire, output); err != nil || !proto.Equal(c.input, output) {
			t.Fatal("legacy calling code hop changed response", err)
		}
	}
}
