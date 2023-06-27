package orderedmap

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var (
	_ yaml.Marshaler   = &OrderedMap[int, any]{}
	_ yaml.Unmarshaler = &OrderedMap[int, any]{}
)

// MarshalYAML implements the yaml.Marshaler interface.
func (om *OrderedMap[K, V]) MarshalYAML() (interface{}, error) { //nolint:funlen
	if om == nil {
		return []byte("null"), nil
	}

	node := yaml.Node{
		Kind: yaml.MappingNode,
	}

	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		key, value := pair.Key, pair.Value

		keyNode := &yaml.Node{}

		// serialize key to yaml, then unserialize it back into the node
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
	if om.list == nil {
		om.initialize(0)
	}

	log.Info().Msgf("UnmarshalYAML: %v", value)

	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("pipeline must contain YAML mapping, has %v", value.Kind)
	}
	for i := 0; i < len(value.Content); i += 2 {
		var key K
		var val V

		if err := value.Content[i].Decode(&key); err != nil {
			return err
		}
		if err := value.Content[i+1].Decode(&val); err != nil {
			return err
		}

		om.Set(key, val)
	}

	return nil
}
