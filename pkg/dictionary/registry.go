package dictionary

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	errorpackage "github.com/siti-nabila/error-package"
)

type Registry struct {
	service     string
	definitions map[string]Definition
}

type yamlCatalog struct {
	Errors map[string]map[string]string `yaml:"errors"`
}

type yamlDefinition struct {
	Code       string
	HTTPStatus int
	Messages   map[string]string
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
		parsed, err := parseYAMLDefinition(service, localKey, raw)
		if err != nil {
			return Registry{}, err
		}
		definition, err := newDefinition(service, localKey, parsed)
		if err != nil {
			return Registry{}, err
		}
		registry.definitions[definition.Key()] = definition
	}
	return registry, nil
}

func parseYAMLDefinition(
	service string,
	localKey string,
	fields map[string]string,
) (yamlDefinition, error) {
	definition := yamlDefinition{Messages: make(map[string]string)}
	for field, value := range fields {
		switch field {
		case "code":
			definition.Code = value
		case "http_status":
			status, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return yamlDefinition{}, fmt.Errorf(
					"load %s.%s error dictionary: http_status must be an integer",
					service,
					strings.TrimSpace(localKey),
				)
			}
			definition.HTTPStatus = status
		default:
			definition.Messages[field] = value
		}
	}
	return definition, nil
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
	return &Error{
		definition: definition,
		messages:   definition.messages,
	}, nil
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
	formatted, formatErr := err.messages.Format(args...)
	if formatErr != nil {
		return nil, fmt.Errorf(
			"create %s error %q: %w",
			registry.service,
			localKey,
			formatErr,
		)
	}
	err.messages = formatted
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

	messages := make(map[errorpackage.LanguageCode]string, len(raw.Messages))
	for language, message := range raw.Messages {
		messages[errorpackage.LanguageCode(language)] = message
	}
	localizedMessages, err := errorpackage.NewLocalizedMessages(messages)
	if err != nil {
		return Definition{}, fmt.Errorf(
			"load %s.%s error dictionary: %w",
			service,
			localKey,
			err,
		)
	}

	return Definition{
		key:        service + "." + localKey,
		code:       strings.TrimSpace(raw.Code),
		httpStatus: raw.HTTPStatus,
		messages:   localizedMessages,
	}, nil
}
