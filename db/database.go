package db

type Encodable interface {
	Encode() ([]byte, error)
}

type Database interface {
	Get(key interface{}) ([]byte, bool)
	Set(key interface{}, value []byte) error
	Del(key interface{}) error
	NewBatch() Batch
}

type CacheKV interface {
	Get(key interface{}) (interface{}, bool)
	Set(key interface{}, value interface{}) error
	Del(key interface{}) error
}

// KeyValueWriter wraps the Put method of a backing data store.
type KeyValueWriter interface {
	// Put inserts the given value into the key-value data store.
	Set(key interface{}, value []byte) error

	// Delete removes the key from the key-value data store.
	Delete(key interface{}) error
}

// Batch is a write-only database that commits changes to its host database
// when Write is called. A batch cannot be used concurrently.
type Batch interface {
	KeyValueWriter

	// ValueSize retrieves the amount of data queued up for writing.
	ValueSize() int

	// Write flushes any accumulated data to disk.
	Write() error

	// Reset resets the batch for reuse.
	Reset()
}
