package dictionary

import (
	"fmt"
	"sort"
	"strings"

	"github.com/siti-nabila/api-contracts/pkg/locale"
)

type FieldErrors map[string][]error

func (fieldErrors FieldErrors) Add(field string, err error) {
	if err == nil {
		return
	}
	fieldErrors[field] = append(fieldErrors[field], err)
}

func (fieldErrors FieldErrors) Merge(other FieldErrors) {
	for field, errors := range other {
		fieldErrors[field] = append(fieldErrors[field], errors...)
	}
}

func (fieldErrors FieldErrors) Empty() bool {
	return len(fieldErrors) == 0
}

func (fieldErrors FieldErrors) Error() string {
	messages := fieldErrors.Messages(locale.DefaultLanguage)
	fields := make([]string, 0, len(messages))
	for field := range messages {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	var builder strings.Builder
	for index, field := range fields {
		if index > 0 {
			builder.WriteString("; ")
		}
		fmt.Fprintf(&builder, "%s: %s", field, strings.Join(messages[field], ", "))
	}
	return builder.String()
}

func (fieldErrors FieldErrors) Messages(language locale.Language) map[string][]string {
	messages := make(map[string][]string)
	appendFieldMessages(messages, "", fieldErrors, language)
	return messages
}

func appendFieldMessages(
	result map[string][]string,
	prefix string,
	fieldErrors FieldErrors,
	language locale.Language,
) {
	for field, errors := range fieldErrors {
		fullField := field
		if prefix != "" {
			fullField = prefix + "." + field
		}
		for _, err := range errors {
			nested, ok := err.(FieldErrors)
			if ok {
				appendFieldMessages(result, fullField, nested, language)
				continue
			}
			localized, ok := err.(interface {
				Message(locale.Language) string
			})
			if ok {
				result[fullField] = append(result[fullField], localized.Message(language))
				continue
			}
			result[fullField] = append(result[fullField], err.Error())
		}
	}
}
