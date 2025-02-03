package internal

import (
	"cmp"
	"fmt"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"reflect"
	"sync"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/internal/jsonflags"
	"github.com/octohelm/courier/pkg/validator/internal/rules"
)

type ValidatorProvider interface {
	Names() []string
	Validator(rule *rules.Rule) (Validator, error)
}

type Validator interface {
	Validate(value jsontext.Value) error
	String() string
}

type PostValidator interface {
	PostValidate(value reflect.Value) error
}

type WithOptional interface {
	Optional() bool
}

type WithDefaultValue interface {
	DefaultValue() jsontext.Value
}

type WithKey interface {
	Key() ValidatorOption
}

type WithElem interface {
	Elem() ValidatorOption
}

var defaultValidators = &validators{
	providers: map[string]ValidatorProvider{},
}

func Register(creator ValidatorProvider) {
	for _, name := range creator.Names() {
		defaultValidators.providers[name] = creator
	}
}

func NewWithStructField(sf *jsonflags.StructField) (Validator, error) {
	opt := ValidatorOption{}

	opt.Type = sf.Type
	opt.String = sf.String
	opt.Optional = cmp.Or(sf.Omitzero, sf.Omitempty)

	if v, ok := sf.Tag.Lookup("validate"); ok {
		opt.Rule = v
	}

	if v, ok := sf.Tag.Lookup("default"); ok {
		if err := opt.SetDefaultValue(v); err != nil {
			return nil, err
		}
	}

	return New(opt)
}

func New(option ValidatorOption) (Validator, error) {
	return defaultValidators.New(option)
}

type ValidatorOption struct {
	Type         reflect.Type
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
	providers map[string]ValidatorProvider
	// map[ValidatorOption]() Validator
	rules sync.Map
}

func (v *validators) New(option ValidatorOption) (Validator, error) {
	get, _ := v.rules.LoadOrStore(option, sync.OnceValues(func() (Validator, error) {
		if option.Rule == "" {
			option.Rule = v.defaultRule(option.Type)
		}

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

		c, ok := v.providers[r.Name]
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

func (v *validators) defaultRule(t reflect.Type) string {
	if t == nil {
		return ""
	}

	if jsonflags.Implements(t, textUnmarshalerType) {
		return "@string"
	}

	switch t {
	case bytesType:
		return "@string"
	default:
		switch t.Kind() {
		case reflect.Ptr:
			return v.defaultRule(t.Elem())
		case reflect.Array:
			return fmt.Sprintf("@slice[%d]", t.Len())
		case reflect.Slice:
			return "@slice?"
		case reflect.Map:
			return "@map?"
		case reflect.Bool:
			return "@bool"
		case reflect.Int:
			return "@int"
		case reflect.Int8:
			return "@int8"
		case reflect.Int16:
			return "@int16"
		case reflect.Int32:
			return "@int32"
		case reflect.Int64:
			return "@int64"
		case reflect.Uint:
			return "@uint"
		case reflect.Uint8:
			return "@uint8"
		case reflect.Uint16:
			return "@uint16"
		case reflect.Uint32:
			return "@uint32"
		case reflect.Uint64:
			return "@uint64"
		case reflect.Float32:
			return "@float"
		case reflect.Float64:
			return "@double"
		case reflect.String:
			return "@string"
		}
	}

	return ""
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
		defaultValue: defaultValue,
	}
}

type wrapValidator struct {
	optional     bool
	underlying   Validator
	defaultValue string
}

func (o *wrapValidator) Optional() bool {
	return o.optional
}

func (o *wrapValidator) Unwrap() Validator {
	return o.underlying
}

func (o *wrapValidator) DefaultValue() jsontext.Value {
	if o.defaultValue == "" {
		return nil
	}
	return jsontext.Value(o.defaultValue)
}

func (o *wrapValidator) String() string {
	if v := o.underlying; v != nil {
		rule := v.String()
		if o.optional {
			return rule + "?"
		}
		return rule
	}

	return ""
}

func (o *wrapValidator) PostValidate(rv reflect.Value) error {
	if post, ok := o.underlying.(PostValidator); ok {
		return post.PostValidate(rv)
	}
	return nil
}

func (o *wrapValidator) Elem() ValidatorOption {
	if post, ok := o.underlying.(WithElem); ok {
		return post.Elem()
	}
	return ValidatorOption{}
}

func (o *wrapValidator) Key() ValidatorOption {
	if post, ok := o.underlying.(WithKey); ok {
		return post.Key()
	}
	return ValidatorOption{}
}

func (o *wrapValidator) Validate(value jsontext.Value) error {
	switch value.Kind() {
	case 'n':
		if o.optional {
			return nil
		}
		return &validatorerrors.ErrMissingRequired{}
	}
	if o.underlying == nil {
		return nil
	}
	return o.underlying.Validate(value)
}
