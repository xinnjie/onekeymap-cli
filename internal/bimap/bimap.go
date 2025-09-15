package bimap

type BiMap[K comparable, V comparable] struct {
	forward map[K]V
	inverse map[V]K
}

// NewBiMap returns a new empty, mutable BiMap.
func NewBiMap[K comparable, V comparable]() *BiMap[K, V] {
	return &BiMap[K, V]{forward: make(map[K]V), inverse: make(map[V]K)}
}

// NewBiMapFromMap returns a new BiMap from a map[K, V].
func NewBiMapFromMap[K comparable, V comparable](forwardMap map[K]V) *BiMap[K, V] {
	biMap := NewBiMap[K, V]()
	for k, v := range forwardMap {
		biMap.Insert(k, v)
	}
	return biMap
}

// Insert inserts a key and value into the BiMap, provided its mutable. Also creates the reverse mapping from value to key.
func (b *BiMap[K, V]) Insert(k K, v V) {
	if _, ok := b.forward[k]; ok {
		delete(b.inverse, b.forward[k])
	}

	b.forward[k] = v
	b.inverse[v] = k
}

// Exists checks whether or not a key exists in the BiMap.
func (b *BiMap[K, V]) Exists(k K) bool {
	_, ok := b.forward[k]
	return ok
}

// ExistsInverse checks whether or not a value exists in the BiMap.
func (b *BiMap[K, V]) ExistsInverse(v V) bool {
	_, ok := b.inverse[v]
	return ok
}

// Get returns the value for a given key in the BiMap and whether or not the element was present.
func (b *BiMap[K, V]) Get(k K) (V, bool) {
	if !b.Exists(k) {
		return *new(V), false
	}
	return b.forward[k], true
}

// GetInverse returns the key for a given value in the BiMap and whether or not the element was present.
func (b *BiMap[K, V]) GetInverse(v V) (K, bool) {
	if !b.ExistsInverse(v) {
		return *new(K), false
	}
	return b.inverse[v], true
}
