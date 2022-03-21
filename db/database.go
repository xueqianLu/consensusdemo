package db

type Encodable interface {
	Encode() ([]byte, error)
}

type Database interface {
	Get(key interface{}) ([]byte, bool)
	Set(key interface{}, value []byte) error
	Del(key interface{}) error
}

type CacheKV interface {
	Get(key interface{}) (interface{}, bool)
	Set(key interface{}, value interface{}) error
	Del(key interface{}) error
}
