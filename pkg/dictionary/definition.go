package dictionary

import (
	"fmt"

	"github.com/siti-nabila/api-contracts/pkg/locale"
	"google.golang.org/grpc/codes"
)

type Definition struct {
	key        string
	code       string
	httpStatus int
	grpcCode   *codes.Code
	english    string
	indonesian string
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

func (definition Definition) GRPCCode() (codes.Code, bool) {
	if definition.grpcCode == nil {
		return codes.OK, false
	}
	return *definition.grpcCode, true
}

func (definition Definition) Message(language locale.Language) string {
	if language == locale.Indonesian && definition.indonesian != "" {
		return definition.indonesian
	}
	return definition.english
}

type Error struct {
	definition Definition
	args       []any
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

func (err *Error) GRPCCode() (codes.Code, bool) {
	if err == nil {
		return codes.OK, false
	}
	return err.definition.GRPCCode()
}

func (err *Error) Message(language locale.Language) string {
	if err == nil {
		return ""
	}
	message := err.definition.Message(language)
	if len(err.args) == 0 {
		return message
	}
	return fmt.Sprintf(message, err.args...)
}
