package internal_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-json-experiment/json/jsontext"

	internalvalidator "github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/courier/pkg/validator/internal/rules"
)

type stubValidator struct {
	validateErr error
	postErr     error
	elem        internalvalidator.ValidatorOption
	key         internalvalidator.ValidatorOption
}

func (s *stubValidator) Validate(value jsontext.Value) error { return s.validateErr }
func (s *stubValidator) String() string                      { return "@stub" }
func (s *stubValidator) PostValidate(reflect.Value) error    { return s.postErr }
func (s *stubValidator) Elem() internalvalidator.ValidatorOption {
	return s.elem
}

func (s *stubValidator) Key() internalvalidator.ValidatorOption {
	return s.key
}

var (
	_ internalvalidator.Validator     = (*stubValidator)(nil)
	_ internalvalidator.PostValidator = (*stubValidator)(nil)
	_ internalvalidator.WithElem      = (*stubValidator)(nil)
	_ internalvalidator.WithKey       = (*stubValidator)(nil)
)

func TestWrapValidatorHelpers(t *testing.T) {
	base := &stubValidator{
		validateErr: errors.New("validate failed"),
		postErr:     errors.New("post failed"),
		elem:        internalvalidator.ValidatorOption{Rule: "@string"},
		key:         internalvalidator.ValidatorOption{Rule: "@int"},
	}

	required := internalvalidator.Required(base)
	if optional, ok := required.(interface{ Optional() bool }); !ok || optional.Optional() {
		t.Fatalf("expected required wrapper to be non-optional")
	}
	if unwrap, ok := required.(interface {
		Unwrap() internalvalidator.Validator
	}); !ok || unwrap.Unwrap() != base {
		t.Fatalf("expected unwrap to return underlying validator")
	}
	if required.String() != "@stub" {
		t.Fatalf("unexpected required string: %q", required.String())
	}
	if err := required.(internalvalidator.PostValidator).PostValidate(reflect.ValueOf("x")); err == nil || err.Error() != "post failed" {
		t.Fatalf("unexpected post validate error: %v", err)
	}
	if elem := required.(interface {
		Elem() internalvalidator.ValidatorOption
	}).Elem(); elem.Rule != "@string" {
		t.Fatalf("unexpected elem option: %#v", elem)
	}
	if key := required.(interface {
		Key() internalvalidator.ValidatorOption
	}).Key(); key.Rule != "@int" {
		t.Fatalf("unexpected key option: %#v", key)
	}
	if err := required.Validate(jsontext.Value("1")); err == nil || err.Error() != "validate failed" {
		t.Fatalf("unexpected validate error: %v", err)
	}
	if err := required.Validate(jsontext.Value("null")); err == nil {
		t.Fatalf("expected missing required error")
	}

	optional := internalvalidator.Optional(base, `"default"`)
	if opt, ok := optional.(interface{ Optional() bool }); !ok || !opt.Optional() {
		t.Fatalf("expected optional wrapper")
	}
	if optional.String() != "@stub?" {
		t.Fatalf("unexpected optional string: %q", optional.String())
	}
	if def, ok := optional.(interface{ DefaultValue() jsontext.Value }); !ok || string(def.DefaultValue()) != `"default"` {
		t.Fatalf("unexpected default value")
	}
	if err := optional.Validate(jsontext.Value("null")); err != nil {
		t.Fatalf("expected optional null to pass, got %v", err)
	}
}

func TestCreateValidatorProviderAndValidator(t *testing.T) {
	provider := internalvalidator.CreateValidatorProvider([]string{"stub"}, func(rule *rules.Rule) (internalvalidator.Validator, error) {
		if rule.Name != "stub" {
			t.Fatalf("unexpected rule name: %s", rule.Name)
		}
		return internalvalidator.CreateValidator("@stub", func(value jsontext.Value) error {
			if string(value) != `"ok"` {
				return errors.New("unexpected value")
			}
			return nil
		}), nil
	})

	names := provider.Names()
	if len(names) != 1 || names[0] != "stub" {
		t.Fatalf("unexpected provider names: %#v", names)
	}

	v, err := provider.Validator(rules.NewRule("stub"))
	if err != nil {
		t.Fatalf("unexpected provider validator error: %v", err)
	}
	if v.String() != "@stub" {
		t.Fatalf("unexpected validator string: %q", v.String())
	}
	if err := v.Validate(jsontext.Value(`"ok"`)); err != nil {
		t.Fatalf("unexpected validator result: %v", err)
	}
	if err := v.Validate(jsontext.Value(`"bad"`)); err == nil || err.Error() != "unexpected value" {
		t.Fatalf("unexpected validator error: %v", err)
	}
}
