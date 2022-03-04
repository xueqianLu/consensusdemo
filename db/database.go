package db

type Database interface {
	Get(key interface{}) (interface{}, bool)
	Set(key interface{}, value interface{}) error
}
