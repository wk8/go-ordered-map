package orderedmap

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

var (
	_ yaml.Marshaler   = &OrderedMap[int, any]{}
	_ yaml.Unmarshaler = &OrderedMap[int, any]{}
)

// MarshalYAML implements the yaml.Marshaler interface.
func (om *OrderedMap[K, V]) MarshalYAML() (interface{}, error) {
	if om == nil {
		return JSONNullBytes(), nil
	}

	node := yaml.Node{
		Kind: yaml.MappingNode,
	}

	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		key, value := pair.Key, pair.Value

		keyNode := &yaml.Node{}

		// serialize key to yaml, then deserialize it back into the node
		// this is a hack to get the correct tag for the key
		if err := keyNode.Encode(key); err != nil {
			return nil, err
		}

		valueNode := &yaml.Node{}
		if err := valueNode.Encode(value); err != nil {
			return nil, err
		}

		node.Content = append(node.Content, keyNode, valueNode)
	}

	return &node, nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (om *OrderedMap[K, V]) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("pipeline must contain YAML mapping, has %v", value.Kind)
	}

	if om.list == nil {
		om.initialize(0, om.disableHTMLEscape)
	}

	for index := 0; index < len(value.Content); index += 2 {
		var key K
		if err := value.Content[index].Decode(&key); err != nil {
			return err
		}

		tValue := reflect.TypeFor[V]()
		tValueKind := tValue.Kind()
		
		var val V

		if tValueKind == reflect.Interface {
			valueNode := value.Content[index+1]
			if bytes.Equal([]byte(valueNode.Value), JSONNullBytes()) {
				om.Set(key, val)
				continue
			}

			om1, err := tryDecodeToOrderedMap(value.Content[index+1].Decode)
			if err == nil {
				om.Set(key, om1.(V))
				continue
			}
		}

		if err := value.Content[index+1].Decode(&val); err != nil {
			return err
		}

		om.Set(key, val)
	}

	return nil
}

// since yaml supports integers and strings as map keys, we need to try both
func tryDecodeToOrderedMap(decoderFunc func(v any) (err error)) (any, error) {
	type newFunc func() any

	possibilities := []newFunc{
		func() any { return New[int, any]() },
		func() any { return New[int32, any]() },
		func() any { return New[int64, any]() },
		func() any { return New[uint, any]() },
		func() any { return New[uint32, any]() },
		func() any { return New[uint64, any]() },
		func() any { return New[string, any]() },
	}

	for _, possibilityFunc := range possibilities {
		possibility := possibilityFunc()
		if err := decoderFunc(possibility); err == nil {
			return possibility, nil
		}
	}

	return nil, errors.New("could not decode to an ordered map")
}
