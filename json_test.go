package orderedmap

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// to test marshalling TextMarshalers and unmarshalling TextUnmarshalers
type marshallable int

func (m marshallable) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("#%d#", m)), nil
}

func (m *marshallable) UnmarshalText(text []byte) error {
	if len(text) < 3 {
		return errors.New("too short")
	}
	if text[0] != '#' || text[len(text)-1] != '#' {
		return errors.New("missing prefix or suffix")
	}

	value, err := strconv.Atoi(string(text[1 : len(text)-1]))
	if err != nil {
		return err
	}

	*m = marshallable(value)
	return nil
}

func TestMarshalJSON(t *testing.T) {
	t.Run("int key", func(t *testing.T) {
		om := New[int, any]()
		om.Set(1, "bar")
		om.Set(7, "baz")
		om.Set(2, 28)
		om.Set(3, 100)
		om.Set(4, "baz")
		om.Set(5, "28")
		om.Set(6, "100")
		om.Set(8, "baz")
		om.Set(8, "baz")
		om.Set(9, "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Quisque auctor augue accumsan mi maximus, quis viverra massa pretium. Phasellus imperdiet sapien a interdum sollicitudin. Duis at commodo lectus, a lacinia sem.")

		b, err := json.Marshal(om)
		assert.NoError(t, err)
		assert.Equal(t, `{"1":"bar","7":"baz","2":28,"3":100,"4":"baz","5":"28","6":"100","8":"baz","9":"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Quisque auctor augue accumsan mi maximus, quis viverra massa pretium. Phasellus imperdiet sapien a interdum sollicitudin. Duis at commodo lectus, a lacinia sem."}`, string(b))
	})

	t.Run("string key", func(t *testing.T) {
		om := New[string, any]()
		om.Set("test", "bar")
		om.Set("abc", true)

		b, err := json.Marshal(om)
		assert.NoError(t, err)
		assert.Equal(t, `{"test":"bar","abc":true}`, string(b))
	})

	t.Run("typed string key", func(t *testing.T) {
		type myString string
		om := New[myString, any]()
		om.Set("test", "bar")
		om.Set("abc", true)

		b, err := json.Marshal(om)
		assert.NoError(t, err)
		assert.Equal(t, `{"test":"bar","abc":true}`, string(b))
	})

	t.Run("typed int key", func(t *testing.T) {
		type myInt uint32
		om := New[myInt, any]()
		om.Set(1, "bar")
		om.Set(7, "baz")
		om.Set(2, 28)
		om.Set(3, 100)
		om.Set(4, "baz")

		b, err := json.Marshal(om)
		assert.NoError(t, err)
		assert.Equal(t, `{"1":"bar","7":"baz","2":28,"3":100,"4":"baz"}`, string(b))
	})

	t.Run("TextMarshaller key", func(t *testing.T) {
		om := New[marshallable, any]()
		om.Set(marshallable(1), "bar")
		om.Set(marshallable(28), true)

		b, err := json.Marshal(om)
		assert.NoError(t, err)
		assert.Equal(t, `{"#1#":"bar","#28#":true}`, string(b))
	})

	t.Run("empty map", func(t *testing.T) {
		om := New[string, any]()

		b, err := json.Marshal(om)
		assert.NoError(t, err)
		assert.Equal(t, `{}`, string(b))
	})

	t.Run("HTML escaping enabled (default)", func(t *testing.T) {
		om := New[marshallable, any]()
		om.Set(marshallable(1), "hello <strong>this is bold</strong>")
		om.Set(marshallable(28), "<?xml version=\"1.0\"?><catalog><book>some book</book></catalog>")

		b, err := jsonMarshal(om, false)
		assert.NoError(t, err)
		assert.Equal(t, `{"#1#":"hello \u003cstrong\u003ethis is bold\u003c/strong\u003e","#28#":"\u003c?xml version=\"1.0\"?\u003e\u003ccatalog\u003e\u003cbook\u003esome book\u003c/book\u003e\u003c/catalog\u003e"}`, string(b))
	})

	t.Run("HTML escaping disabled", func(t *testing.T) {
		om := New[marshallable, any](WithDisableHTMLEscape[marshallable, any]())
		om.Set(marshallable(1), "hello <strong>this is bold</strong>")
		om.Set(marshallable(28), "<?xml version=\"1.0\"?><catalog><book>some book</book></catalog>")

		b, err := jsonMarshal(om, true /* we need to disable HTML escaping here also */)
		assert.NoError(t, err)
		assert.Equal(t, `{"#1#":"hello <strong>this is bold</strong>","#28#":"<?xml version=\"1.0\"?><catalog><book>some book</book></catalog>"}`, string(b))
	})
}

func TestUnmarshallJSON(t *testing.T) {
	t.Run("int key", func(t *testing.T) {
		data := `{"1":"bar","7":"baz","2":28,"3":100,"4":"baz","5":"28","6":"100","8":"baz"}`

		om := New[int, any]()
		require.NoError(t, json.Unmarshal([]byte(data), &om))

		assertOrderedPairsEqual(t, om,
			[]int{1, 7, 2, 3, 4, 5, 6, 8},
			[]any{"bar", "baz", float64(28), float64(100), "baz", "28", "100", "baz"})
	})

	t.Run("string key", func(t *testing.T) {
		data := `{"test":"bar","abc":true}`

		om := New[string, any]()
		require.NoError(t, json.Unmarshal([]byte(data), &om))

		assertOrderedPairsEqual(t, om,
			[]string{"test", "abc"},
			[]any{"bar", true})
	})

	t.Run("typed string key", func(t *testing.T) {
		data := `{"test":"bar","abc":true}`

		type myString string
		om := New[myString, any]()
		require.NoError(t, json.Unmarshal([]byte(data), &om))

		assertOrderedPairsEqual(t, om,
			[]myString{"test", "abc"},
			[]any{"bar", true})
	})

	t.Run("typed int key", func(t *testing.T) {
		data := `{"1":"bar","7":"baz","2":28,"3":100,"4":"baz","5":"28","6":"100","8":"baz"}`

		type myInt uint32
		om := New[myInt, any]()
		require.NoError(t, json.Unmarshal([]byte(data), &om))

		assertOrderedPairsEqual(t, om,
			[]myInt{1, 7, 2, 3, 4, 5, 6, 8},
			[]any{"bar", "baz", float64(28), float64(100), "baz", "28", "100", "baz"})
	})

	t.Run("TextUnmarshaler key", func(t *testing.T) {
		data := `{"#1#":"bar","#28#":true}`

		om := New[marshallable, any]()
		require.NoError(t, json.Unmarshal([]byte(data), &om))

		assertOrderedPairsEqual(t, om,
			[]marshallable{1, 28},
			[]any{"bar", true})
	})

	t.Run("when fed with an input that's not an object", func(t *testing.T) {
		for _, data := range []string{"true", `["foo"]`, "42", `"foo"`} {
			om := New[int, any]()
			require.Error(t, json.Unmarshal([]byte(data), &om))
		}
	})

	t.Run("empty map", func(t *testing.T) {
		data := `{}`

		om := New[int, any]()
		require.NoError(t, json.Unmarshal([]byte(data), &om))

		assertLenEqual(t, om, 0)
	})
}

func TestUnmarshalJSONNullRoundTrip(t *testing.T) {
	var original OrderedMap[string, int]
	data, err := json.Marshal(&original)
	require.NoError(t, err)
	require.Equal(t, "null", string(data))

	var decoded OrderedMap[string, int]
	require.NoError(t, json.Unmarshal(data, &decoded))
	assertLenEqual(t, &decoded, 0)
	assert.Nil(t, decoded.Oldest())
	assert.Nil(t, decoded.Newest())
	decoded.Set("new", 1)
	assertOrderedPairsEqual(t, &decoded, []string{"new"}, []int{1})
}

func TestUnmarshalJSONNullClears(t *testing.T) {
	for _, testCase := range []struct {
		name string
		make func() *OrderedMap[string, int]
	}{
		{"zero", func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} }},
		{"empty", func() *OrderedMap[string, int] { return New[string, int]() }},
		{"populated", func() *OrderedMap[string, int] {
			return New[string, int](WithInitialData(Pair[string, int]{Key: "old", Value: 1}))
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			for _, data := range []string{"null", " \t\r\nnull\n\r\t "} {
				om := testCase.make()
				require.NoError(t, json.Unmarshal([]byte(data), om))
				assertLenEqual(t, om, 0)
				assert.Nil(t, om.Oldest())
				assert.Nil(t, om.Newest())
				_, present := om.Get("old")
				assert.False(t, present)
				om.Set("temporary", 2)
				require.NoError(t, json.Unmarshal([]byte("null"), om))
				require.NoError(t, json.Unmarshal([]byte(`{"second":2,"first":1}`), om))
				assertOrderedPairsEqual(t, om, []string{"second", "first"}, []int{2, 1})
			}
		})
	}
}

func TestUnmarshalJSONNullPreservesHTMLOption(t *testing.T) {
	for _, om := range []*OrderedMap[string, string]{
		{disableHTMLEscape: true},
		New[string, string](WithDisableHTMLEscape[string, string](),
			WithInitialData(Pair[string, string]{Key: "old", Value: "value"})),
	} {
		require.NoError(t, json.Unmarshal([]byte("null"), om))
		om.Set("<key>", "<value>&")
		data, err := jsonMarshal(om, true)
		require.NoError(t, err)
		assert.Equal(t, `{"<key>":"<value>&"}`, string(data))
		require.NoError(t, json.Unmarshal([]byte("null"), om))
		require.NoError(t, json.Unmarshal([]byte(`{"<new>":"<value>&"}`), om))
		data, err = jsonMarshal(om, true)
		require.NoError(t, err)
		assert.Equal(t, `{"<new>":"<value>&"}`, string(data))
	}
}

func TestUnmarshalJSONNullPointer(t *testing.T) {
	populated := New[string, int](WithInitialData(Pair[string, int]{Key: "old", Value: 1}))
	for _, om := range []*OrderedMap[string, int]{nil, populated} {
		require.NoError(t, json.Unmarshal([]byte("null"), &om))
		assert.Nil(t, om)
	}
}

func TestUnmarshalJSONObjectsPreserveOrder(t *testing.T) {
	var zero OrderedMap[string, int]
	require.NoError(t, json.Unmarshal([]byte(`{"second":2,"first":1,"second":3}`), &zero))
	assertOrderedPairsEqual(t, &zero, []string{"second", "first"}, []int{3, 1})

	om := New[string, int]()
	om.Set("old", 1)
	om.Set("keep", 2)
	require.NoError(t, json.Unmarshal([]byte(`{"keep":3,"new":4,"keep":5,"last":6}`), om))
	assertOrderedPairsEqual(t, om, []string{"old", "keep", "new", "last"}, []int{1, 5, 4, 6})
	require.NoError(t, json.Unmarshal([]byte(`{}`), om))
	assertOrderedPairsEqual(t, om, []string{"old", "keep", "new", "last"}, []int{1, 5, 4, 6})
}

func TestUnmarshalJSONRejectsNonObjects(t *testing.T) {
	for _, data := range []string{"true", "false", "[]", `["foo"]`, "42", `"foo"`} {
		t.Run(data, func(t *testing.T) {
			om := New[string, int](WithInitialData(Pair[string, int]{Key: "old", Value: 1}))
			require.Error(t, json.Unmarshal([]byte(data), om))
			assertOrderedPairsEqual(t, om, []string{"old"}, []int{1})
		})
	}
}

func TestUnmarshalJSONRejectsInvalidJSON(t *testing.T) {
	for _, data := range []string{"", "nul", "Null", "nullx", "null {}", "null null", "null\v", `{"old":2,}`} {
		t.Run(data, func(t *testing.T) {
			om := New[string, int](WithInitialData(Pair[string, int]{Key: "old", Value: 1}))
			err := json.Unmarshal([]byte(data), om)
			var syntaxError *json.SyntaxError
			require.ErrorAs(t, err, &syntaxError)
			assertOrderedPairsEqual(t, om, []string{"old"}, []int{1})
		})
	}
}

func assertJSONIntegerKey[K comparable](t *testing.T, expected K) {
	t.Helper()
	t.Run(fmt.Sprintf("%T", expected), func(t *testing.T) {
		var om OrderedMap[K, string]
		require.NoError(t, json.Unmarshal([]byte(`{"7":"seven"}`), &om))
		assertOrderedPairsEqual(t, &om, []K{expected}, []string{"seven"})
	})
}

func TestUnmarshalJSONIntegerKeys(t *testing.T) {
	assertJSONIntegerKey(t, int(7))
	assertJSONIntegerKey(t, int8(7))
	assertJSONIntegerKey(t, int16(7))
	assertJSONIntegerKey(t, int32(7))
	assertJSONIntegerKey(t, int64(7))
	assertJSONIntegerKey(t, uint(7))
	assertJSONIntegerKey(t, uint8(7))
	assertJSONIntegerKey(t, uint16(7))
	assertJSONIntegerKey(t, uint32(7))
	assertJSONIntegerKey(t, uint64(7))
	type namedInt int16
	type namedUint uint16
	assertJSONIntegerKey(t, namedInt(7))
	assertJSONIntegerKey(t, namedUint(7))
}

type jsonErrorKey struct{}

func (*jsonErrorKey) UnmarshalText([]byte) error {
	return strconv.ErrSyntax
}

func TestUnmarshalJSONKeyErrors(t *testing.T) {
	t.Run("TextUnmarshaler", func(t *testing.T) {
		var om OrderedMap[jsonErrorKey, int]
		err := json.Unmarshal([]byte(`{"key":1}`), &om)
		require.ErrorIs(t, err, strconv.ErrSyntax)
		assert.Equal(t, strconv.ErrSyntax.Error(), err.Error())
	})
	t.Run("unsupported key", func(t *testing.T) {
		var om OrderedMap[bool, int]
		require.EqualError(t, json.Unmarshal([]byte(`{"true":1}`), &om), "unsupported key type: bool")
	})
	t.Run("invalid UTF-8", func(t *testing.T) {
		var om OrderedMap[string, int]
		require.EqualError(t, json.Unmarshal([]byte("{\"\xff\":1}"), &om), "not a valid UTF-8 string (at position 0): \xff")
		type namedString string
		var named OrderedMap[namedString, int]
		require.EqualError(t, json.Unmarshal([]byte("{\"\xff\":1}"), &named), "not a valid UTF-8 string (at position 0): \xff")
	})
	t.Run("integer overflow", func(t *testing.T) {
		var om OrderedMap[int8, int]
		err := json.Unmarshal([]byte(`{"128":1}`), &om)
		var typeError *json.UnmarshalTypeError
		require.ErrorAs(t, err, &typeError)
		assert.Equal(t, "number 128", typeError.Value)
		assert.Equal(t, "int8", typeError.Type.String())
		require.EqualError(t, err, "json: cannot unmarshal number 128 into Go value of type int8")
	})
	t.Run("named integer overflow", func(t *testing.T) {
		type namedInt int8
		var om OrderedMap[namedInt, int]
		err := json.Unmarshal([]byte(`{"128":1}`), &om)
		var typeError *json.UnmarshalTypeError
		require.ErrorAs(t, err, &typeError)
		assert.Equal(t, "number 128", typeError.Value)
		assert.Equal(t, "orderedmap.namedInt", typeError.Type.String())
	})
	t.Run("negative unsigned integer", func(t *testing.T) {
		var om OrderedMap[uint, int]
		err := json.Unmarshal([]byte(`{"-1":1}`), &om)
		var typeError *json.UnmarshalTypeError
		require.ErrorAs(t, err, &typeError)
		assert.Equal(t, "number -1", typeError.Value)
		assert.Equal(t, "uint", typeError.Type.String())
	})
	t.Run("invalid integer", func(t *testing.T) {
		var om OrderedMap[int, int]
		err := json.Unmarshal([]byte(`{"abc":1}`), &om)
		var syntaxError *json.SyntaxError
		require.ErrorAs(t, err, &syntaxError)
		require.EqualError(t, err, "invalid character 'a' looking for beginning of value")
	})
}

func TestUnmarshalJSONNullDoesNotDecodeKeys(t *testing.T) {
	failing := New[jsonErrorKey, int](WithInitialData(Pair[jsonErrorKey, int]{Key: jsonErrorKey{}, Value: 1}))
	require.NoError(t, json.Unmarshal([]byte("null"), failing))
	assertLenEqual(t, failing, 0)
	failing.Set(jsonErrorKey{}, 2)
	assertOrderedPairsEqual(t, failing, []jsonErrorKey{{}}, []int{2})
	unsupported := New[bool, int](WithInitialData(Pair[bool, int]{Key: true, Value: 1}))
	require.NoError(t, json.Unmarshal([]byte("null"), unsupported))
	assertLenEqual(t, unsupported, 0)
	unsupported.Set(false, 2)
	assertOrderedPairsEqual(t, unsupported, []bool{false}, []int{2})
}

func TestUnmarshalJSONValueErrorKeepsEarlierUpdates(t *testing.T) {
	om := New[string, int](WithInitialData(Pair[string, int]{Key: "old", Value: 1}))
	err := json.Unmarshal([]byte(`{"first":2,"bad":"invalid","last":3}`), om)
	var typeError *json.UnmarshalTypeError
	require.ErrorAs(t, err, &typeError)
	assert.Equal(t, "string", typeError.Value)
	assert.Equal(t, "int", typeError.Type.String())
	assertOrderedPairsEqual(t, om, []string{"old", "first"}, []int{1, 2})

	var nested OrderedMap[string, OrderedMap[string, int]]
	err = json.Unmarshal([]byte(`{"nested":{"bad":"invalid"}}`), &nested)
	require.ErrorAs(t, err, &typeError)
	assert.Equal(t, "string", typeError.Value)
	assert.Equal(t, "int", typeError.Type.String())
	assertLenEqual(t, &nested, 0)
}

// const specialCharacters = "\\\\/\"\b\f\n\r\t\x00\uffff\ufffd世界\u007f\u00ff\U0010FFFF"
const specialCharacters = "\uffff\ufffd世界\u007f\u00ff\U0010FFFF"

func TestJSONSpecialCharacters(t *testing.T) {
	baselineMap := map[string]any{specialCharacters: specialCharacters}
	baselineData, err := json.Marshal(baselineMap)
	require.NoError(t, err) // baseline proves this key is supported by official json library
	t.Logf("specialCharacters: %#v as []rune:%v", specialCharacters, []rune(specialCharacters))
	t.Logf("baseline json data: %s", baselineData)

	t.Run("marshal special characters", func(t *testing.T) {
		om := New[string, any]()
		om.Set(specialCharacters, specialCharacters)
		b, err := json.Marshal(om)
		require.NoError(t, err)
		require.Equal(t, baselineData, b)

		type myString string
		om2 := New[myString, myString]()
		om2.Set(specialCharacters, specialCharacters)
		b, err = json.Marshal(om2)
		require.NoError(t, err)
		require.Equal(t, baselineData, b)
	})

	t.Run("unmarshall special characters", func(t *testing.T) {
		om := New[string, any]()
		require.NoError(t, json.Unmarshal(baselineData, &om))
		assertOrderedPairsEqual(t, om,
			[]string{specialCharacters},
			[]any{specialCharacters})

		type myString string
		om2 := New[myString, myString]()
		require.NoError(t, json.Unmarshal(baselineData, &om2))
		assertOrderedPairsEqual(t, om2,
			[]myString{specialCharacters},
			[]myString{specialCharacters})
	})
}

// to test structs that have nested map fields
type nestedMaps struct {
	X int                                                               `json:"x" yaml:"x"`
	M *OrderedMap[string, []*OrderedMap[int, *OrderedMap[string, any]]] `json:"m" yaml:"m"`
}

func TestJSONRoundTrip(t *testing.T) {
	for _, testCase := range []struct {
		name            string
		input           string
		targetFactory   func() any
		isPrettyPrinted bool
	}{
		{
			name: "",
			input: `{
    "x": 28,
    "m": {
        "foo": [
            {
                "12": {
                    "i": 12,
                    "b": true,
                    "n": null,
                    "m": {
                        "a": "b",
                        "c": 28
                    }
                },
                "28": {
                    "a": false,
                    "b": [
                        1,
                        2,
                        3
                    ]
                }
            },
            {
                "3": {
                    "c": null,
                    "d": 87
                },
                "4": {
                    "e": true
                },
                "5": {
                    "f": 4,
                    "g": 5,
                    "h": 6
                }
            }
        ],
        "bar": [
            {
                "5": {
                    "foo": "bar"
                }
            }
        ]
    }
}`,
			targetFactory:   func() any { return &nestedMaps{} },
			isPrettyPrinted: true,
		},
		{
			name:          "with UTF-8 special chars in key",
			input:         `{"�":0}`,
			targetFactory: func() any { return &OrderedMap[string, int]{} },
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			target := testCase.targetFactory()

			require.NoError(t, json.Unmarshal([]byte(testCase.input), target))

			var (
				out []byte
				err error
			)
			if testCase.isPrettyPrinted {
				out, err = json.MarshalIndent(target, "", "    ")
			} else {
				out, err = json.Marshal(target)
			}

			if assert.NoError(t, err) {
				assert.Equal(t, strings.TrimSpace(testCase.input), string(out))
			}
		})
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	om := New[int, any]()
	om.Set(1, "bar")
	om.Set(7, "baz")
	om.Set(2, 28)
	om.Set(3, 100)
	om.Set(4, "baz")
	om.Set(5, "28")
	om.Set(6, "100")
	om.Set(8, "baz")
	om.Set(8, "baz")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(om)
	}
}
