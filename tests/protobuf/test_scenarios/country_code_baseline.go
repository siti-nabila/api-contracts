package test_scenarios

import (
	"google.golang.org/protobuf/types/descriptorpb"
)

// Keep historical descriptor fixtures frozen. Separately tested additive Me
// permissions and calling-code fields are removed before exact baseline checks.
func beforeCountryCode(file *descriptorpb.FileDescriptorProto) *descriptorpb.FileDescriptorProto {
	out := beforeMePermissions(file)
	if out.GetName() != "profile/v1/profile.payload.proto" && out.GetName() != "user/v1/user.payload.proto" {
		return out
	}
	for _, message := range out.MessageType {
		fields := message.Field[:0]
		for _, field := range message.Field {
			if field.GetName() != "country_code" {
				fields = append(fields, field)
			}
		}
		message.Field = fields
		oneofs := message.OneofDecl[:0]
		for _, oneof := range message.OneofDecl {
			if oneof.GetName() != "_country_code" {
				oneofs = append(oneofs, oneof)
			}
		}
		message.OneofDecl = oneofs
	}
	return out
}
