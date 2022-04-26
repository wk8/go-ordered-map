// Package orderedmap implements an ordered map, i.e. a map that also keeps track of
// the order in which keys were inserted.
//
// All operations are constant-time.
//
// Github repo: https://github.com/wk8/go-ordered-map
//
package orderedmap

import (
	"container/list"
	"fmt"
)

type Pair struct {
	Key   interface{}
	Value interface{}

	element *list.Element
}

type OrderedMap struct {
	pairs map[interface{}]*Pair
	list  *list.List
}

// New creates a new OrderedMap.
func New() *OrderedMap {
	return &OrderedMap{
		pairs: make(map[interface{}]*Pair),
		list:  list.New(),
	}
}

// Get looks for the given key, and returns the value associated with it,
// or nil if not found. The boolean it returns says whether the key is present in the map.
func (om *OrderedMap) Get(key interface{}) (interface{}, bool) {
	if pair, present := om.pairs[key]; present {
		return pair.Value, present
	}
	return nil, false
}
func (om *OrderedMap) Load(key interface{}) (interface{}, bool) {
	return om.Get(key)
}

// GetPair looks for the given key, and returns the pair associated with it,
// or nil if not found. The Pair struct can then be used to iterate over the ordered map
// from that point, either forward or backward.
func (om *OrderedMap) GetPair(key interface{}) *Pair {
	return om.pairs[key]
}

// Set sets the key-value pair, and returns what `Get` would have returned
// on that key prior to the call to `Set`.
func (om *OrderedMap) Set(key interface{}, value interface{}) (interface{}, bool) {
	if pair, present := om.pairs[key]; present {
		oldValue := pair.Value
		pair.Value = value
		return oldValue, true
	}

	pair := &Pair{
		Key:   key,
		Value: value,
	}
	pair.element = om.list.PushBack(pair)
	om.pairs[key] = pair

	return nil, false
}
func (om *OrderedMap) Store(key interface{}, value interface{}) (interface{}, bool) {
	return om.Set(key, value)
}

// Delete removes the key-value pair, and returns what `Get` would have returned
// on that key prior to the call to `Delete`.
func (om *OrderedMap) Delete(key interface{}) (interface{}, bool) {
	if pair, present := om.pairs[key]; present {
		om.list.Remove(pair.element)
		delete(om.pairs, key)
		return pair.Value, true
	}
	return nil, false
}

// Len returns the length of the ordered map.
func (om *OrderedMap) Len() int {
	return len(om.pairs)
}

// Oldest returns a pointer to the oldest pair. It's meant to be used to iterate on the ordered map's
// pairs from the oldest to the newest, e.g.:
// for pair := orderedMap.Oldest(); pair != nil; pair = pair.Next() { fmt.Printf("%v => %v\n", pair.Key, pair.Value) }
func (om *OrderedMap) Oldest() *Pair {
	return listElementToPair(om.list.Front())
}

// Newest returns a pointer to the newest pair. It's meant to be used to iterate on the ordered map's
// pairs from the newest to the oldest, e.g.:
// for pair := orderedMap.Oldest(); pair != nil; pair = pair.Next() { fmt.Printf("%v => %v\n", pair.Key, pair.Value) }
func (om *OrderedMap) Newest() *Pair {
	return listElementToPair(om.list.Back())
}

// Next returns a pointer to the next pair.
func (p *Pair) Next() *Pair {
	return listElementToPair(p.element.Next())
}

// Previous returns a pointer to the previous pair.
func (p *Pair) Prev() *Pair {
	return listElementToPair(p.element.Prev())
}

func listElementToPair(element *list.Element) *Pair {
	if element == nil {
		return nil
	}
	return element.Value.(*Pair)
}

func (om *OrderedMap) MoveAfter(key interface{}, mark_key interface{}) error {
	var e, mark *list.Element
	if pair, present := om.pairs[key]; present {
		e = pair.element
	} else {
		return fmt.Errorf("error: key %v not found", key)
	}
	if pair, present := om.pairs[mark_key]; present {
		mark = pair.element
	} else {
		return fmt.Errorf("error: mark_key %v not found", mark_key)
	}
	om.list.MoveAfter(e, mark)
	return nil
}
func (om *OrderedMap) MoveBefore(key interface{}, mark_key interface{}) error {
	var e, mark *list.Element
	if pair, present := om.pairs[key]; present {
		e = pair.element
	} else {
		return fmt.Errorf("error: key %v not found", key)
	}
	if pair, present := om.pairs[mark_key]; present {
		mark = pair.element
	} else {
		return fmt.Errorf("error: mark_key %v not found", mark_key)
	}
	om.list.MoveBefore(e, mark)
	return nil
}
func (om *OrderedMap) MoveToBack(key interface{}) error {
	var e *list.Element
	if pair, present := om.pairs[key]; present {
		e = pair.element
	} else {
		return fmt.Errorf("error: key %v not found", key)
	}
	om.list.MoveToBack(e)
	return nil
}
func (om *OrderedMap) MoveToFront(key interface{}) error {
	var e *list.Element
	if pair, present := om.pairs[key]; present {
		e = pair.element
	} else {
		return fmt.Errorf("error: key %v not found", key)
	}
	om.list.MoveToFront(e)
	return nil
}
