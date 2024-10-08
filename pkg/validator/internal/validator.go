package internal

import (
	"bytes"
	"fmt"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"reflect"
	"sync"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/validator/internal/rules"
)

type ValidatorCreator interface {
	Names() []string
	Validator(rule *rules.Rule) (Validator, error)
}

type Validator interface {
	Validate(value jsontext.Value) (jsontext.Value, error)
}

type PostValidator interface {
	PostValidate(value reflect.Value) error
}

var defaultValidators = &validators{
	creators: map[string]ValidatorCreator{},
}

func Register(creator ValidatorCreator) {
	for _, name := range creator.Names() {
		defaultValidators.creators[name] = creator
	}
}

func New(option ValidatorOption) (Validator, error) {
	return defaultValidators.New(option)
}

type ValidatorOption struct {
	Rule         string
	String       bool
	Optional     bool
	DefaultValue string
}

func (o *ValidatorOption) SetDefaultValue(v string) error {
	if o.String {
		raw, err := jsontext.AppendQuote(nil, []byte(v))
		if err != nil {
			return err
		}
		o.DefaultValue = string(raw)
	} else {
		o.DefaultValue = v
	}
	return nil
}

type validators struct {
	creators map[string]ValidatorCreator
	// map[ValidatorOption]() Validator
	rules sync.Map
}

func (v *validators) New(option ValidatorOption) (Validator, error) {
	get, _ := v.rules.LoadOrStore(option, sync.OnceValues(func() (Validator, error) {
		if option.Rule == "" {
			if option.Optional {
				return Optional(nil, option.DefaultValue), nil
			}
			return Required(nil), nil
		}

		r, err := rules.ParseRuleString(option.Rule)
		if err != nil {
			return nil, err
		}

		if r.Optional {
			option.Optional = r.Optional
		}

		if r.DefaultValue != nil {
			if err := option.SetDefaultValue(string(r.DefaultValue)); err != nil {
				return nil, err
			}
		}

		c, ok := v.creators[r.Name]
		if !ok {
			return nil, fmt.Errorf("unknown supported rule %s", string(r.Bytes()))
		}

		validator, err := c.Validator(r)
		if err != nil {
			return nil, err
		}

		if option.Optional {
			return Optional(validator, option.DefaultValue), nil
		}

		return Required(validator), nil
	}))

	return get.(func() (Validator, error))()
}

func Required(v Validator) Validator {
	return &wrapValidator{
		underlying: v,
	}
}

func Optional(v Validator, defaultValue string) Validator {
	return &wrapValidator{
		underlying:   v,
		optional:     true,
		defaultValue: jsontext.Value(defaultValue),
	}
}

type wrapValidator struct {
	optional     bool
	underlying   Validator
	defaultValue jsontext.Value
}

func (o *wrapValidator) Validate(value jsontext.Value) (jsontext.Value, error) {
	if !o.optional {
		switch value.Kind() {
		case 'n':
			return value, &validatorerrors.MissingRequired{}
		}
	} else {
		if len(o.defaultValue) > 0 {
			switch value.Kind() {
			case 'n':
				value = o.defaultValue
			case '"':
				if bytes.Equal(value, []byte("\"\"")) {
					value = o.defaultValue
				}
			case '0':
				if bytes.Equal(value, []byte("0")) {
					value = o.defaultValue
				}
			case 'b':
				if bytes.Equal(value, []byte("false")) {
					value = o.defaultValue
				}
			}
		}
	}
	if o.underlying == nil {
		return value, nil
	}
	return o.underlying.Validate(value)
}
