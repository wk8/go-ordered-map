package orderedmap

// Adapted from https://github.com/dvyukov/go-fuzz-corpus/blob/c42c1b2/json/json.go

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/dvyukov/go-fuzz-corpus/fuzz"
)

func FuzzMarshalling(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		for _, ctor := range []func() any{
			func() any { return nil },
			func() any { return new([]any) },
			func() any { return &OrderedMap[string, string]{} },
			func() any { return &OrderedMap[string, any]{} },
			func() any { return new(S) },
		} {
			v := ctor()
			if json.Unmarshal(data, v) != nil {
				continue
			}
			if s, ok := v.(*S); ok {
				if len(s.P) == 0 {
					s.P = []byte(`""`)
				}
			}
			data1, err := json.Marshal(v)
			if err != nil {
				panic(err)
			}
			v1 := ctor()
			if json.Unmarshal(data1, v1) != nil {
				continue
			}
			if s, ok := v.(*S); ok {
				// Some additional escaping happens with P.
				s.P = nil
				v1.(*S).P = nil
			}
			if !fuzz.DeepEqual(v, v1) {
				fmt.Printf("v0: %#v\n", v)
				fmt.Printf("v1: %#v\n", v1)
				panic("not equal")
			}
		}
	})
}

type S struct {
	A int    `json:",omitempty"`
	B string `json:"B1,omitempty"`
	C float64
	D bool
	E uint8
	F []byte
	G any
	H OrderedMap[string, any]
	I OrderedMap[string, string]
	J []any
	K []string
	L S1
	M *S1
	N *int
	O **int
	P json.RawMessage
	Q Marshaller
	R int `json:"-"`
	S int `json:",string"`
}

type S1 struct {
	A int
	B string
}

type Marshaller struct {
	v string
}

func (m *Marshaller) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.v)
}

func (m *Marshaller) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.v)
}
