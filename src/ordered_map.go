package main

// Ordered Map Data Structure

// a map that preserves the order of its keys
type OrderedMap[K comparable, V any] struct {
	data map[K]V
	keys []K
}

// constructor for an ordered map
func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		data: map[K]V{},
		keys: []K{},
	}
}

// adds a key-value pair to the map, or adjusts an existing one
func (omap *OrderedMap[K, V]) Set(key K, value V) {
	if !omap.Has(key) {
		omap.keys = append(omap.keys, key)
	}
	omap.data[key] = value
}

// determines if a key is in the map
func (omap *OrderedMap[K, V]) Has(key K) bool {
	_, isKeyExists := omap.data[key]
	return isKeyExists
}

// provides a way to iterate over the map's keys in the correct order
func (omap *OrderedMap[K, V]) Iterator() func() (*int, *K, V) {
	// iterate over the map's keys
	var keys = omap.keys
	i := 0

	return func() (_ *int, _ *K, _ V) {
		if i > len(keys)-1 {
			return
		}

		key := keys[i]
		i++

		return &[]int{i - 1}[0], &key, omap.data[key]
	}
}
