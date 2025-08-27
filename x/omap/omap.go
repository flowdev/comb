// Package omap implements a very simple ordered map with just the absolute
// minimum features for our purpose.
// It's not very performant of whatever, but it's nice to use and easy to maintain.
package omap

import (
	"cmp"
	"iter"
)

// OrderedMap implements a map ordered by the keys.
type OrderedMap[K cmp.Ordered, V any] struct {
	m map[K]V
	s []K
}

// New returns a new ordered map with space for size elements.
func New[K cmp.Ordered, V any](size int) *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		m: make(map[K]V, size),
		s: make([]K, 0, size),
	}
}

func (om *OrderedMap[K, V]) Get(k K) (V, bool) {
	v, ok := om.m[k]
	return v, ok
}

func (om *OrderedMap[K, V]) GetFirst() (K, V) {
	var k K
	var v V

	if len(om.s) == 0 {
		return k, v
	}
	k = om.s[0]
	v, _ = om.m[k]
	return k, v
}

func (om *OrderedMap[K, V]) ReplaceFirst(k K, v V) {
	if len(om.s) == 0 {
		om.Add(k, v)
		return
	}
	if cmp.Compare(k, om.s[0]) != 0 {
		delete(om.m, om.s[0])
	}
	om.m[k] = v
	om.s[0] = k
	n := len(om.s)
	for i := 0; i < n-1; i++ { // keep it sorted
		switch c := cmp.Compare(k, om.s[i+1]); c {
		case -1:
			// we are already sorted
			return
		case 0: // same key? we have one less value!
			om.s = append(om.s[:i+1], om.s[i+2:]...)
			return
		case 1:
			om.s[i], om.s[i+1] = om.s[i+1], om.s[i]
		}
	}
}

// Add puts the value v into the map under the key k.
// It will overwrite any existing value.
func (om *OrderedMap[K, V]) Add(k K, v V) {
	if _, ok := om.m[k]; !ok {
		om.s = append(om.s, k)
		n := len(om.s)
	FOR:
		for i := n - 1; i > 0; i-- { // keep it sorted
			switch c := cmp.Compare(k, om.s[i-1]); c {
			case -1:
				om.s[i], om.s[i-1] = om.s[i-1], om.s[i]
			case 0: // same key? we have one less value!
				om.s = om.s[:n-1]
				break FOR
			case 1:
				// we are already sorted
				break FOR
			}
		}
	}
	om.m[k] = v
}

func (om *OrderedMap[K, V]) Exists(k K) bool {
	_, ok := om.m[k]
	return ok
}

func (om *OrderedMap[K, V]) Len() int {
	return len(om.s)
}

// All returns an iterator that can be used to iterate
// like over normal maps (with a for range loop).
func (om *OrderedMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(key K, value V) bool) {
		for _, k := range om.s {
			if !yield(k, om.m[k]) {
				return
			}
		}
	}
}

// Build adds key k and value v to the ordered map and returns the map itself.
// This allows elegant building of maps.
func (om *OrderedMap[K, V]) Build(k K, v V) *OrderedMap[K, V] {
	om.Add(k, v)
	return om
}
