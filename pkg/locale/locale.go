package locale

import (
	"context"
	"strings"

	errorpackage "github.com/siti-nabila/error-package"
)

type Language string

const (
	English    Language = "en"
	Indonesian Language = "id"
	Chinese    Language = "zh"

	DefaultLanguage = English
	MetadataKey     = "x-language"
	HTTPHeader      = "Accept-Language"
)

type contextKey struct{}

func Parse(value string) Language {
	first := strings.TrimSpace(strings.Split(value, ",")[0])
	first = strings.TrimSpace(strings.Split(first, ";")[0])
	parsed, err := errorpackage.ParseLanguageCode(first)
	if err != nil {
		return DefaultLanguage
	}
	return Language(parsed)
}

func NewContext(ctx context.Context, language string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, contextKey{}, Parse(language))
}

func FromContext(ctx context.Context) Language {
	if ctx == nil {
		return DefaultLanguage
	}
	language, ok := ctx.Value(contextKey{}).(Language)
	if !ok {
		return DefaultLanguage
	}
	return language
}
