package orderedmap

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"

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
		return []byte("null"), nil
	}

	writer := jwriter.Writer{}
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
		writer.Raw(json.Marshal(pair.Value)) //nolint:errchkjson
	}

	writer.RawByte('}')

	return dumpWriter(&writer)
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
		om.initialize(0)
	}

	return jsonparser.ObjectEach(
		data,
		func(keyData []byte, valueData []byte, dataType jsonparser.ValueType, offset int) error {
			if dataType == jsonparser.String {
				// jsonparser removes the enclosing quotes; we need to restore them to make a valid JSON
				valueData = data[offset-len(valueData)-2 : offset]
			}

			var key K
			var value V

			if typedKeyPointer, ok := any(&key).(encoding.TextUnmarshaler); ok {
				// pointer receiver
				if err := typedKeyPointer.UnmarshalText(keyData); err != nil {
					return err
				}
			} else {
				keyAlreadyUnmarshalled := false
				switch typedKey := any(key).(type) {
				case string:
					keyData = quoteString(keyData)
				case encoding.TextUnmarshaler:
					// not a pointer receiver
					if err := typedKey.UnmarshalText(keyData); err != nil {
						return err
					}
					keyAlreadyUnmarshalled = true
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				default:

					// this switch takes care of wrapper types around primitive types, such as
					// type myType string
					switch reflect.TypeOf(key).Kind() {
					case reflect.String:
						keyData = quoteString(keyData)
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
						reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					default:
						return fmt.Errorf("unsupported key type: %T", typedKey)
					}
				}

				if !keyAlreadyUnmarshalled {
					if err := json.Unmarshal(keyData, &key); err != nil {
						return err
					}
				}
			}

			if err := json.Unmarshal(valueData, &value); err != nil {
				return err
			}

			om.Set(key, value)
			return nil
		})
}

func quoteString(data []byte) []byte {
	withQuotes := make([]byte, len(data)+2) //nolint:gomnd
	copy(withQuotes[1:], data)
	withQuotes[0] = '"'
	withQuotes[len(data)+1] = '"'
	return withQuotes
}
