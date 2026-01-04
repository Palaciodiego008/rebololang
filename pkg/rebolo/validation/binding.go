package validation

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Bind binds request data to a struct
// Supports form data, JSON, and query parameters
func Bind(r *http.Request, v interface{}) error {
	if v == nil {
		return errors.New("bind target cannot be nil")
	}

	// Check if it's JSON request
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return bindJSON(r, v)
	}

	// Parse form data (including multipart)
	if err := r.ParseForm(); err != nil {
		return err
	}

	return bindForm(r, v)
}

// bindJSON binds JSON request body to struct
func bindJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	return decoder.Decode(v)
}

// bindForm binds form data to struct
func bindForm(r *http.Request, v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return errors.New("bind target must be a pointer")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("bind target must be a struct pointer")
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := typ.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get form tag or use lowercase field name
		tag := typeField.Tag.Get("form")
		if tag == "" {
			tag = strings.ToLower(typeField.Name)
		}

		// Skip if tag is "-"
		if tag == "-" {
			continue
		}

		// Get value from form
		formValue := r.FormValue(tag)
		if formValue == "" {
			continue
		}

		// Set value based on field type
		if err := setField(field, formValue); err != nil {
			return err
		}
	}

	return nil
}

// setField sets a struct field value from string
func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			// Also accept "on" for checkboxes
			boolVal = value == "on" || value == "1"
		}
		field.SetBool(boolVal)

	default:
		return errors.New("unsupported field type: " + field.Kind().String())
	}

	return nil
}

// Validator interface for custom validation
type Validator interface {
	Validate() error
}

// Validate validates a struct if it implements Validator interface
func Validate(v interface{}) error {
	if validator, ok := v.(Validator); ok {
		return validator.Validate()
	}
	return nil
}

// BindAndValidate binds and validates in one step
func BindAndValidate(r *http.Request, v interface{}) error {
	if err := Bind(r, v); err != nil {
		return err
	}
	return Validate(v)
}

