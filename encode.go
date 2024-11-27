package hl7

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

var (
	ErrInvalidBooleanValue = errors.New("fixedlength: invalid boolean value")
	ErrInvalidIntValue     = errors.New("fixedlength: invalid int value")
	ErrInvalidFloatValue   = errors.New("fixedlength: invalid float value")
	ErrUnsupportedKind     = errors.New("fixedlength: unsupported kind")
)

// setFieldValue sets the value for a struct field using reflection.
// If the value is an empty string and the field is a pointer, it sets the pointer to nil.
func setFieldValue(field reflect.Value, value string) error {
	// Handle pointers by allocating and dereferencing them.
	if field.Kind() == reflect.Ptr {
		if value == "" {
			// If the value is an empty string, set the pointer to nil.
			field.Set(reflect.Zero(field.Type()))
			return nil
		}

		if field.IsNil() {
			// Allocate new value and set it to the pointer.
			field.Set(reflect.New(field.Type().Elem()))
		}

		// Dereference the pointer.
		field = field.Elem()
	}

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
