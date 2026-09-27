package test_scenarios

import (
	"errors"
	"sync"
	"testing"

	"github.com/siti-nabila/api-contracts/pkg/dictionary"
	"github.com/siti-nabila/api-contracts/pkg/dictionary/common"
	"github.com/siti-nabila/api-contracts/pkg/locale"
	"github.com/siti-nabila/api-contracts/tests/dictionary/fixtures"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
)

func All() []testutils.Scenario {
	return []testutils.Scenario{
		{
			Name: "loads valid yaml and resolves localized definition",
			Run:  loadValidCatalog,
		},
		{
			Name: "common registry exposes localized endpoint not found error",
			Run:  exposeEndpointNotFoundError,
		},
		{
			Name: "rejects invalid locale field",
			Run:  rejectUnknownField,
		},
		{
			Name: "rejects grpc code in yaml catalog",
			Run:  rejectGRPCCode,
		},
		{
			Name: "compares errors by service-qualified key",
			Run:  compareByQualifiedKey,
		},
		{
			Name: "localizes concurrent requests without shared language state",
			Run:  localizeConcurrently,
		},
	}
}

func exposeEndpointNotFoundError(t *testing.T) {
	err := common.ErrEndpointNotFound

	if err.Key() != "common.endpoint_not_found" {
		t.Errorf("Key() = %q, want common.endpoint_not_found", err.Key())
	}
	if err.Code() != "NF" {
		t.Errorf("Code() = %q, want NF", err.Code())
	}
	if err.HTTPStatus() != 404 {
		t.Errorf("HTTPStatus() = %d, want 404", err.HTTPStatus())
	}
	if message := err.Message(locale.English); message != "Endpoint not found." {
		t.Errorf("Message(en) = %q, want %q", message, "Endpoint not found.")
	}
	if message := err.Message(locale.Indonesian); message != "Endpoint tidak ditemukan." {
		t.Errorf(
			"Message(id) = %q, want %q",
			message,
			"Endpoint tidak ditemukan.",
		)
	}
}

func loadValidCatalog(t *testing.T) {
	registry, err := dictionary.LoadYAML("auth", []byte(fixtures.ValidCatalog))
	if err != nil {
		t.Fatalf("LoadYAML() error = %v", err)
	}

	definition, exists := registry.Lookup("auth.missing")
	if !exists {
		t.Fatal("Lookup() did not find auth.missing")
	}
	if definition.Code() != "NF" {
		t.Errorf("Code() = %q, want NF", definition.Code())
	}
	if definition.HTTPStatus() != 404 {
		t.Errorf("HTTPStatus() = %d, want 404", definition.HTTPStatus())
	}
	if message := definition.Message(locale.Indonesian); message != "Tidak ditemukan." {
		t.Errorf("Message(id) = %q, want %q", message, "Tidak ditemukan.")
	}
	if message := definition.Message(locale.Chinese); message != "未找到。" {
		t.Errorf("Message(zh) = %q, want %q", message, "未找到。")
	}
	if message := definition.Message(locale.Language("zh-CN")); message != "未找到。" {
		t.Errorf("Message(zh-CN) = %q, want base-language fallback", message)
	}
	if _, overridden := definition.GRPCCode(); overridden {
		t.Error("GRPCCode() reports an override without a YAML transport mapping")
	}
}

func rejectUnknownField(t *testing.T) {
	_, err := dictionary.LoadYAML("auth", []byte(fixtures.UnknownFieldCatalog))
	if err == nil {
		t.Fatal("LoadYAML() error = nil, want unknown-field error")
	}
}

func rejectGRPCCode(t *testing.T) {
	_, err := dictionary.LoadYAML("auth", []byte(fixtures.GRPCCodeCatalog))
	if err == nil {
		t.Fatal("LoadYAML() error = nil, want grpc_code to be rejected")
	}
}

func compareByQualifiedKey(t *testing.T) {
	authRegistry, err := dictionary.LoadYAML("auth", []byte(fixtures.ValidCatalog))
	if err != nil {
		t.Fatalf("load auth registry: %v", err)
	}
	orderRegistry, err := dictionary.LoadYAML("order", []byte(fixtures.ValidCatalog))
	if err != nil {
		t.Fatalf("load order registry: %v", err)
	}

	authError := authRegistry.MustNew("missing")
	sameAuthError := authRegistry.MustNew("missing")
	orderError := orderRegistry.MustNew("missing")

	if !errors.Is(authError, sameAuthError) {
		t.Error("errors.Is() = false for the same qualified key")
	}
	if errors.Is(authError, orderError) {
		t.Error("errors.Is() = true for different service-qualified keys")
	}
}

func localizeConcurrently(t *testing.T) {
	registry, err := dictionary.LoadYAML("auth", []byte(fixtures.ValidCatalog))
	if err != nil {
		t.Fatalf("LoadYAML() error = %v", err)
	}
	applicationError := registry.MustNew("missing")

	type expectation struct {
		language locale.Language
		message  string
	}
	expectations := []expectation{
		{language: locale.English, message: "Not found."},
		{language: locale.Indonesian, message: "Tidak ditemukan."},
	}

	var waitGroup sync.WaitGroup
	failures := make(chan string, 200)
	for index := 0; index < 100; index++ {
		for _, expected := range expectations {
			expected := expected
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				if message := applicationError.Message(expected.language); message != expected.message {
					failures <- message
				}
			}()
		}
	}
	waitGroup.Wait()
	close(failures)

	for message := range failures {
		t.Errorf("concurrent Message() = %q", message)
	}
}
