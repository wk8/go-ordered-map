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
			func() any { return &OrderedMap[string, string]{} },
			func() any { return &OrderedMap[string, any]{} },
			func() any { return new(S) },
		} {
			v := ctor()
			if json.Unmarshal(data, v) != nil {
				continue
			}
			data1, err := json.Marshal(v)
			if err != nil {
				panic(err)
			}

			if s, ok := v.(*S); ok {
				if len(s.P) == 0 {
					s.P = []byte(`""`)
				}
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
				// Skip false positives due to linked list initialization
				switch v := v.(type) {
				case *OrderedMap[string, string]:
					if l := v1.(*OrderedMap[string, string]).Len(); l == 0 && l == v.Len() {
						continue
					}
				case *OrderedMap[string, any]:
					if l := v1.(*OrderedMap[string, any]).Len(); l == 0 && l == v.Len() {
						continue
					}
				case *S:
					if l := v1.(*S).H.Len(); l == 0 && l == v.H.Len() {
						if ll := v1.(*S).I.Len(); ll == 0 && ll == v.I.Len() {
							continue
						}
					}
				default:
					panic(fmt.Sprintf("unhandled %T", v))
				}
				fmt.Printf("v0: %#v\n", v)
				fmt.Printf("v1: %#v\n", v1)
				panic("not equal")
			}
		}
	})
}

type S struct {
	H OrderedMap[int, any]
	I OrderedMap[int, string]
	P json.RawMessage
}
