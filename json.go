package orderedmap

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"unicode/utf8"

	"github.com/buger/jsonparser"
	"github.com/mailru/easyjson/jwriter"
)

var (
	_ json.Marshaler   = &OrderedMap[int, any]{}
	_ json.Unmarshaler = &OrderedMap[int, any]{}
)

// MarshalJSON implements the json.Marshaler interface.
func (om *OrderedMap[K, V]) MarshalJSON() ([]byte, error) { //nolint:funlen
	if om == nil || om.list == nil {
		return JSONNullBytes(), nil
	}

	writer := jwriter.Writer{
		NoEscapeHTML: om.disableHTMLEscape,
	}
	writer.RawByte('{')

	for pair, firstIteration := om.Oldest(), true; pair != nil; pair = pair.Next() {
		if firstIteration {
			firstIteration = false
		} else {
			writer.RawByte(',')
		}

		switch key := any(pair.Key).(type) {
		case string:
			writer.String(key)
		case encoding.TextMarshaler:
			writer.RawByte('"')
			writer.Raw(key.MarshalText())
			writer.RawByte('"')
		case int:
			writer.IntStr(key)
		case int8:
			writer.Int8Str(key)
		case int16:
			writer.Int16Str(key)
		case int32:
			writer.Int32Str(key)
		case int64:
			writer.Int64Str(key)
		case uint:
			writer.UintStr(key)
		case uint8:
			writer.Uint8Str(key)
		case uint16:
			writer.Uint16Str(key)
		case uint32:
			writer.Uint32Str(key)
		case uint64:
			writer.Uint64Str(key)
		default:

			// this switch takes care of wrapper types around primitive types, such as
			// type myType string
			switch keyValue := reflect.ValueOf(key); keyValue.Type().Kind() {
			case reflect.String:
				writer.String(keyValue.String())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				writer.Int64Str(keyValue.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				writer.Uint64Str(keyValue.Uint())
			default:
				return nil, fmt.Errorf("unsupported key type: %T", key)
			}
		}

		writer.RawByte(':')
		// the error is checked at the end of the function
		writer.Raw(jsonMarshal(pair.Value, om.disableHTMLEscape))
	}

	writer.RawByte('}')

	return dumpWriter(&writer)
}

func jsonMarshal(t interface{}, disableHTMLEscape bool) ([]byte, error) {
	if disableHTMLEscape {
		buffer := &bytes.Buffer{}
		encoder := json.NewEncoder(buffer)
		encoder.SetEscapeHTML(false)
		err := encoder.Encode(t)
		// Encode() adds an extra newline, strip it off to guarantee same behavior as json.Marshal
		return bytes.TrimRight(buffer.Bytes(), "\n"), err
	}
	return json.Marshal(t)
}

func dumpWriter(writer *jwriter.Writer) ([]byte, error) {
	if writer.Error != nil {
		return nil, writer.Error
	}

	var buf bytes.Buffer
	buf.Grow(writer.Size())
	if _, err := writer.DumpTo(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (om *OrderedMap[K, V]) UnmarshalJSON(data []byte) error {
	if om.list == nil {
		om.initialize(0, om.disableHTMLEscape)
	}

	return jsonparser.ObjectEach(
		data,
		func(keyData []byte, valueData []byte, dataType jsonparser.ValueType, offset int) error {
			if dataType == jsonparser.String {
				// jsonparser removes the enclosing quotes; we need to restore them to make a valid JSON
				valueData = data[offset-len(valueData)-2 : offset]
			}

			var key K

			switch typedKey := any(&key).(type) {
			case *string:
				s, err := decodeUTF8(keyData)
				if err != nil {
					return err
				}
				*typedKey = s
			case encoding.TextUnmarshaler:
				if err := typedKey.UnmarshalText(keyData); err != nil {
					return err
				}
			case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
				if err := json.Unmarshal(keyData, typedKey); err != nil {
					return err
				}
			default:
				// this switch takes care of wrapper types around primitive types, such as
				// type myType string
				switch reflect.TypeOf(key).Kind() {
				case reflect.String:
					s, err := decodeUTF8(keyData)
					if err != nil {
						return err
					}

					convertedKeyData := reflect.ValueOf(s).Convert(reflect.TypeOf(key))
					reflect.ValueOf(&key).Elem().Set(convertedKeyData)
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if err := json.Unmarshal(keyData, &key); err != nil {
						return err
					}
				default:
					return fmt.Errorf("unsupported key type: %T", key)
				}
			}

			tValue := reflect.TypeFor[V]()
			tValueKind := tValue.Kind()

			var value V

			// in the case of an `any`
			// the value may be nil, we should represent it as such instead of an empty map
			// otherwise
			// try to unmarshall into an ordered map before falling back to
			// what json.Unmarshal would regularly do, which, when the value is a map, it would use map[string]any
			if tValueKind == reflect.Interface {
				if bytes.Equal(valueData, JSONNullBytes()) {
					om.Set(key, value)
					return nil
				}
				om1 := New[string, any]()
				if err := json.Unmarshal(valueData, &om1); err == nil {
					om.Set(key, any(om1).(V))
					return nil
				}
			}

			// fallback to json.Unmarshal of V
			if err := json.Unmarshal(valueData, &value); err != nil {
				return err
			}

			om.Set(key, value)
			return nil
		})
}

func decodeUTF8(input []byte) (string, error) {
	remaining, offset := input, 0
	runes := make([]rune, 0, len(remaining))

	for len(remaining) > 0 {
		r, size := utf8.DecodeRune(remaining)
		if r == utf8.RuneError && size <= 1 {
			return "", fmt.Errorf("not a valid UTF-8 string (at position %d): %s", offset, string(input))
		}

		runes = append(runes, r)
		remaining = remaining[size:]
		offset += size
	}

	return string(runes), nil
}
