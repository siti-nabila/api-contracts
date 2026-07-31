package dictionary

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
	"google.golang.org/grpc/codes"
)

type Registry struct {
	service     string
	definitions map[string]Definition
}

type yamlCatalog struct {
	Errors map[string]yamlDefinition `yaml:"errors"`
}

type yamlDefinition struct {
	Code       string `yaml:"code"`
	HTTPStatus int    `yaml:"http_status"`
	GRPCCode   string `yaml:"grpc_code"`
	English    string `yaml:"en"`
	Indonesian string `yaml:"id"`
}

func LoadYAML(service string, data []byte) (Registry, error) {
	service = strings.TrimSpace(service)
	if service == "" {
		return Registry{}, fmt.Errorf("load error dictionary: service must not be empty")
	}

	var catalog yamlCatalog
	if err := yaml.UnmarshalWithOptions(data, &catalog, yaml.Strict()); err != nil {
		return Registry{}, fmt.Errorf("load %s error dictionary: %w", service, err)
	}
	if len(catalog.Errors) == 0 {
		return Registry{}, fmt.Errorf("load %s error dictionary: errors must not be empty", service)
	}

	registry := Registry{
		service:     service,
		definitions: make(map[string]Definition, len(catalog.Errors)),
	}
	for localKey, raw := range catalog.Errors {
		definition, err := newDefinition(service, localKey, raw)
		if err != nil {
			return Registry{}, err
		}
		registry.definitions[definition.Key()] = definition
	}
	return registry, nil
}

func MustLoadYAML(service string, data []byte) Registry {
	registry, err := LoadYAML(service, data)
	if err != nil {
		panic(err)
	}
	return registry
}

func (registry Registry) Lookup(key string) (Definition, bool) {
	definition, exists := registry.definitions[key]
	return definition, exists
}

func (registry Registry) New(localKey string) (*Error, error) {
	definition, exists := registry.Lookup(registry.fullKey(localKey))
	if !exists {
		return nil, fmt.Errorf(
			"create %s error: key %q is not registered",
			registry.service,
			localKey,
		)
	}
	return &Error{definition: definition}, nil
}

func (registry Registry) MustNew(localKey string) *Error {
	err, createErr := registry.New(localKey)
	if createErr != nil {
		panic(createErr)
	}
	return err
}

func (registry Registry) Newf(localKey string, args ...any) (*Error, error) {
	err, createErr := registry.New(localKey)
	if createErr != nil {
		return nil, createErr
	}
	err.args = append([]any(nil), args...)
	return err, nil
}

func (registry Registry) MustNewf(localKey string, args ...any) *Error {
	err, createErr := registry.Newf(localKey, args...)
	if createErr != nil {
		panic(createErr)
	}
	return err
}

func (registry Registry) fullKey(localKey string) string {
	return registry.service + "." + strings.TrimSpace(localKey)
}

func newDefinition(
	service string,
	localKey string,
	raw yamlDefinition,
) (Definition, error) {
	localKey = strings.TrimSpace(localKey)
	if localKey == "" {
		return Definition{}, fmt.Errorf(
			"load %s error dictionary: error key must not be empty",
			service,
		)
	}
	if strings.TrimSpace(raw.English) == "" {
		return Definition{}, fmt.Errorf(
			"load %s.%s error dictionary: en must not be empty",
			service,
			localKey,
		)
	}
	if raw.HTTPStatus < 0 || raw.HTTPStatus > 599 {
		return Definition{}, fmt.Errorf(
			"load %s.%s error dictionary: http_status must be between 0 and 599",
			service,
			localKey,
		)
	}
	if raw.Code != "" && raw.HTTPStatus == 0 {
		return Definition{}, fmt.Errorf(
			"load %s.%s error dictionary: http_status is required when code is set",
			service,
			localKey,
		)
	}

	var grpcCode *codes.Code
	if strings.TrimSpace(raw.GRPCCode) != "" {
		parsed, err := parseGRPCCode(raw.GRPCCode)
		if err != nil {
			return Definition{}, fmt.Errorf(
				"load %s.%s error dictionary: %w",
				service,
				localKey,
				err,
			)
		}
		grpcCode = &parsed
	}

	return Definition{
		key:        service + "." + localKey,
		code:       strings.TrimSpace(raw.Code),
		httpStatus: raw.HTTPStatus,
		grpcCode:   grpcCode,
		english:    raw.English,
		indonesian: raw.Indonesian,
	}, nil
}

func parseGRPCCode(value string) (codes.Code, error) {
	grpcCodes := map[string]codes.Code{
		"CANCELED":            codes.Canceled,
		"UNKNOWN":             codes.Unknown,
		"INVALID_ARGUMENT":    codes.InvalidArgument,
		"DEADLINE_EXCEEDED":   codes.DeadlineExceeded,
		"NOT_FOUND":           codes.NotFound,
		"ALREADY_EXISTS":      codes.AlreadyExists,
		"PERMISSION_DENIED":   codes.PermissionDenied,
		"RESOURCE_EXHAUSTED":  codes.ResourceExhausted,
		"FAILED_PRECONDITION": codes.FailedPrecondition,
		"ABORTED":             codes.Aborted,
		"OUT_OF_RANGE":        codes.OutOfRange,
		"UNIMPLEMENTED":       codes.Unimplemented,
		"INTERNAL":            codes.Internal,
		"UNAVAILABLE":         codes.Unavailable,
		"DATA_LOSS":           codes.DataLoss,
		"UNAUTHENTICATED":     codes.Unauthenticated,
	}
	normalized := strings.ToUpper(strings.TrimSpace(value))
	code, exists := grpcCodes[normalized]
	if !exists {
		return codes.Unknown, fmt.Errorf("unsupported grpc_code %q", value)
	}
	return code, nil
}
