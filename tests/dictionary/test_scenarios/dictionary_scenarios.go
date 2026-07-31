package test_scenarios

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/siti-nabila/api-contracts/pkg/dictionary"
	"github.com/siti-nabila/api-contracts/pkg/locale"
	"github.com/siti-nabila/api-contracts/tests/dictionary/fixtures"
	"github.com/siti-nabila/api-contracts/tests/shared/testutils"
	"google.golang.org/grpc/codes"
)

func All() []testutils.Scenario {
	return []testutils.Scenario{
		{
			Name: "loads valid yaml and resolves localized definition",
			Run:  loadValidCatalog,
		},
		{
			Name: "rejects unknown yaml field",
			Run:  rejectUnknownField,
		},
		{
			Name: "uses optional grpc code override",
			Run:  useGRPCCodeOverride,
		},
		{
			Name: "rejects unsupported grpc code",
			Run:  rejectUnsupportedGRPCCode,
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
	if _, overridden := definition.GRPCCode(); overridden {
		t.Error("GRPCCode() reports an override for omitted grpc_code")
	}
}

func rejectUnknownField(t *testing.T) {
	_, err := dictionary.LoadYAML("auth", []byte(fixtures.UnknownFieldCatalog))
	if err == nil {
		t.Fatal("LoadYAML() error = nil, want unknown-field error")
	}
}

func useGRPCCodeOverride(t *testing.T) {
	registry, err := dictionary.LoadYAML("auth", []byte(fixtures.OverrideCatalog))
	if err != nil {
		t.Fatalf("LoadYAML() error = %v", err)
	}
	definition, exists := registry.Lookup("auth.state")
	if !exists {
		t.Fatal("Lookup() did not find auth.state")
	}

	grpcCode, overridden := definition.GRPCCode()
	if !overridden {
		t.Fatal("GRPCCode() override = false, want true")
	}
	if grpcCode != codes.FailedPrecondition {
		t.Errorf("GRPCCode() = %s, want %s", grpcCode, codes.FailedPrecondition)
	}
}

func rejectUnsupportedGRPCCode(t *testing.T) {
	_, err := dictionary.LoadYAML("auth", []byte(fixtures.InvalidGRPCCodeCatalog))
	if err == nil {
		t.Fatal("LoadYAML() error = nil, want unsupported grpc code error")
	}
	if !strings.Contains(err.Error(), "unsupported grpc_code") {
		t.Errorf("LoadYAML() error = %q, want unsupported grpc_code", err)
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
