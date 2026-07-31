package locale

import (
	"context"
	"strings"
)

type Language string

const (
	English    Language = "en"
	Indonesian Language = "id"

	DefaultLanguage = English
	MetadataKey     = "x-language"
	HTTPHeader      = "Accept-Language"
)

type contextKey struct{}

func Parse(value string) Language {
	first := strings.TrimSpace(strings.Split(value, ",")[0])
	first = strings.TrimSpace(strings.Split(first, ";")[0])
	first = strings.ToLower(first)

	switch {
	case first == string(Indonesian), strings.HasPrefix(first, "id-"):
		return Indonesian
	case first == string(English), strings.HasPrefix(first, "en-"):
		return English
	default:
		return DefaultLanguage
	}
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
