package hl7

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

var (
	ErrInvalidBooleanValue = errors.New("hl7: invalid boolean value")
	ErrInvalidIntValue     = errors.New("hl7: invalid int value")
	ErrInvalidFloatValue   = errors.New("hl7: invalid float value")
	ErrInvalidTimeValue    = errors.New("hl7: invalid time value")
	ErrInvalidValue        = errors.New("hl7: invalid value")
	ErrUnsupportedKind     = errors.New("hl7: unsupported kind")
)

// setFieldValue sets the value for a struct field using reflection.
// If the value is an empty string and the field is a pointer, it sets the pointer to nil.
func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return errors.Join(ErrInvalidIntValue, err)
		}
		field.SetInt(intVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return errors.Join(ErrInvalidFloatValue, err)
		}
		field.SetFloat(floatVal)

	case reflect.String:
		field.SetString(value)

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Join(ErrInvalidBooleanValue, err)
		}
		field.SetBool(boolVal)

	case reflect.Struct:
		switch field.Type() {
		case reflect.TypeOf(time.Time{}):
			parsedTime, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return fmt.Errorf("%w: invalid time value %q", ErrInvalidTimeValue, value)
			}
			field.Set(reflect.ValueOf(parsedTime))
		default:
			return fmt.Errorf("%w: unsupported struct type %s", ErrUnsupportedKind, field.Type())
		}

	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}

		if err := setFieldValue(field.Elem(), value); err != nil {
			field.Set(reflect.Zero(field.Type()))
			return err
		}

	default:
		if implementsUnmarshaler(field) {
			// no need to check if the field is addressable because implementsUnmarshaler already does that
			um, ok := field.Addr().Interface().(Unmarshaler)
			// but one can never be too sure
			if !ok {
				return fmt.Errorf("%w: %s", ErrUnsupportedKind, field.Kind())
			}

			return um.Unmarshal([]byte(value))
		}

		return fmt.Errorf("%w: %s", ErrUnsupportedKind, field.Kind())
	}

	return nil
}
