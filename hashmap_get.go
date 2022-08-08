package hashmap

// Get retrieves an element from the map under given hash key.
// Using interface{} adds a performance penalty.
// Please consider using GetUintKey or GetStringKey instead.
func (m *HashMap[Key, Value]) Get(key Key) (value interface{}, ok bool) {
	hash := getKeyHash(key)
	data, element := m.indexElement(hash)
	if data == nil {
		return nil, false
	}

	// inline HashMap.searchItem()
	for element != nil {
		if element.keyHash == hash && element.key == key {
			return element.Value(), true
		}

		if element.keyHash > hash {
			return nil, false
		}

		element = element.Next()
	}
	return nil, false
}

// GetOrInsert returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *HashMap[Key, Value]) GetOrInsert(key Key, value Value) (actual Value, loaded bool) {
	hash := getKeyHash(key)
	var newElement *ListElement[Key, Value]

	for {
		data, element := m.indexElement(hash)
		if data == nil {
			m.allocate(DefaultSize)
			continue
		}

		for element != nil {
			if element.keyHash == hash && element.key == key {
				actual = element.Value()
				return actual, true
			}

			if element.keyHash > hash {
				break
			}

			element = element.Next()
		}

		if newElement == nil { // allocate only once
			newElement = &ListElement[Key, Value]{
				key:     key,
				keyHash: hash,
			}
			newElement.value.Store(&value)
		}

		if m.insertListElement(newElement, false) {
			return value, false
		}
	}
}
