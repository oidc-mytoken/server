package model

// KeyValue is a type for a key-value pair
type KeyValue struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// KeyValues is a slice of KeyValue
type KeyValues []KeyValue
