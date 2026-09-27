package dictionary

import (
	"github.com/siti-nabila/api-contracts/pkg/locale"
	errorpackage "github.com/siti-nabila/error-package"
	"google.golang.org/grpc/codes"
)

type Definition struct {
	key        string
	code       string
	httpStatus int
	messages   errorpackage.LocalizedMessages
}

func (definition Definition) Key() string {
	return definition.key
}

func (definition Definition) Code() string {
	return definition.code
}

func (definition Definition) HTTPStatus() int {
	return definition.httpStatus
}

// GRPCCode is retained for source compatibility. gRPC mappings are no longer
// loaded from dictionary YAML and therefore this method never reports an
// override.
// Deprecated: configure mappings with grpcerror.NewEncoder.
func (definition Definition) GRPCCode() (codes.Code, bool) {
	return codes.OK, false
}

func (definition Definition) Message(language locale.Language) string {
	return definition.messages.Message(errorpackage.LanguageCode(language))
}

type Error struct {
	definition Definition
	messages   errorpackage.LocalizedMessages
}

func (err *Error) Error() string {
	if err == nil {
		return ""
	}
	return err.Message(locale.DefaultLanguage)
}

func (err *Error) Is(target error) bool {
	other, ok := target.(*Error)
	if !ok || err == nil || other == nil {
		return false
	}
	return err.Key() == other.Key()
}

func (err *Error) Key() string {
	if err == nil {
		return ""
	}
	return err.definition.Key()
}

func (err *Error) Code() string {
	if err == nil {
		return ""
	}
	return err.definition.Code()
}

func (err *Error) HTTPStatus() int {
	if err == nil {
		return 0
	}
	return err.definition.HTTPStatus()
}

// GRPCCode is retained for source compatibility.
// Deprecated: configure mappings with grpcerror.NewEncoder.
func (err *Error) GRPCCode() (codes.Code, bool) {
	return codes.OK, false
}

func (err *Error) Message(language locale.Language) string {
	if err == nil {
		return ""
	}
	return err.messages.Message(errorpackage.LanguageCode(language))
}
