package telegram

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// EncodeForm encodes a struct to map[string]string for form submission.
// Basic types (int, bool, etc.) are converted to string.
// Complex types (structs, slices, pointers) are JSON encoded.
func ToFormValues(v any) map[string]string {
	result := make(map[string]string)
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return result
	}
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)
		// Skip unexported fields
		if !fieldVal.CanInterface() {
			continue
		}
		// Get json tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		// Parse json tag (handle omitempty etc.)
		tagParts := strings.Split(jsonTag, ",")
		key := tagParts[0]
		if key == "" {
			key = field.Name
		}
		value := fieldVal.Interface()
		// Check if value is zero/empty
		isZero := isZeroValue(fieldVal)
		hasOmitEmpty := len(tagParts) > 1 && tagParts[1] == "omitempty"
		// Skip omitempty fields with zero values
		if hasOmitEmpty && isZero {
			continue
		}
		// JSON stringify complex types
		if isComplexType(fieldVal) {
			jsonBytes, err := json.Marshal(value)
			if err == nil {
				result[key] = string(jsonBytes)
			}
		} else {
			// Convert basic types to string
			result[key] = fmt.Sprintf("%v", value)
		}
	}
	return result
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Struct:
		return v.IsZero()
	default:
		return false
	}
}

func isComplexType(v reflect.Value) bool {
	kind := v.Kind()
	switch kind {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Struct, reflect.Array, reflect.Interface:
		// Skip string (which is a slice of bytes)
		if v.Type().String() == "string" {
			return false
		}
		return true
	default:
		return false
	}
}
