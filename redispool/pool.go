package redispool

import (
	"fmt"
	"reflect"
	"time"
	"unsafe"

	"github.com/gomodule/redigo/redis"
)

type RedisPool struct {
	pool *redis.Pool
}

func NewRedisPool(conn, dbNum, password string) (*RedisPool, error) {
	pool := &redis.Pool{
		MaxIdle:     50, //最大空闲连接数
		MaxActive:   0,  //若为0，则活跃数没有限制
		Wait:        true,
		IdleTimeout: 30 * time.Second, //最大空闲连接时间
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", conn)
			if err != nil {
				return nil, err
			}
			if len(password) > 0 {
				// 设置密码
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			// 选择db
			c.Do("SELECT", dbNum)
			return c, nil
		},
	}
	return &RedisPool{pool: pool}, nil
}

/**
设置key,value数据
*/
func (s *RedisPool) Set(key, value string) error {
	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value)

	if err != nil {
		return err
	}
	return nil
}

/**
设置key,value数据
*/
func (s *RedisPool) SetBytes(key string, value []byte) error {
	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value)

	if err != nil {
		return err
	}
	return nil
}

/**
设置key的过期时间
*/
func (s *RedisPool) SetKvAndExp(key, value string, expire int) error {
	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value, "EX", expire)

	if err != nil {
		return err
	}
	return nil
}

func (s *RedisPool) SetKvInt(key string, value int) error {
	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value)

	if err != nil {
		return err
	}
	return nil
}

/**
根据key获取对应数据
*/
func (s *RedisPool) Get(key string) string {
	conn := s.pool.Get()
	defer conn.Close()

	value, err := redis.String(conn.Do("GET", key))

	if err != nil {
		fmt.Println("redis get failed:", err)
	}
	return value
}

/**
根据key获取过期时间
*/
func (s *RedisPool) GetExp(key string) int {
	conn := s.pool.Get()
	defer conn.Close()

	value, err := redis.Int(conn.Do("TTL", key))

	if err != nil {
		fmt.Println("redis get failed:", err)
	}
	return value
}

/**
根据key获取对应数据
*/
func (s *RedisPool) GetInt(key string) int {
	conn := s.pool.Get()
	defer conn.Close()

	value, err := redis.Int(conn.Do("GET", key))

	if err != nil {
		fmt.Println("redis get failed:", err)
	}
	return value
}

/**
根据key获取对应数据
*/
func (s *RedisPool) GetBytes(key string) []byte {
	conn := s.pool.Get()
	defer conn.Close()

	value, err := conn.Do("GET", key)

	if err != nil {
		fmt.Println("redis get failed:", err)
	}
	return value.([]byte)
}

/**
判断key是否存在
*/
func (s *RedisPool) IsKeyExists(key string) bool {
	conn := s.pool.Get()
	defer conn.Close()

	is_key_exit, _ := redis.Bool(conn.Do("EXISTS", key))

	return is_key_exit
}

/**
删除key
*/
func (s *RedisPool) Del(key string) bool {
	conn := s.pool.Get()
	defer conn.Close()

	is_key_delete, err := redis.Bool(conn.Do("DEL", key))

	if err != nil {
		fmt.Println("error:", err)
	}
	return is_key_delete
}

/**
对象转换成json后进行存储
*/
func (s *RedisPool) Setnx(key string, value []byte) error {
	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.Do("SETNX", key, value)

	if err != nil {
		return err
	}
	return nil
}

func (s *RedisPool) LpushByte(key string, value []byte) error {
	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.Do("lpush", key, value)
	if err != nil {

		return err
	}

	return nil
}

func (s *RedisPool) LPopByte(key string) ([]byte, error) {
	conn := s.pool.Get()
	defer conn.Close()

	v, err := redis.Bytes(conn.Do("lpop", key))
	if err != nil {

		return nil, err
	}

	return v, nil
}

func (s *RedisPool) Lpush(key string, value ...int) error {
	conn := s.pool.Get()
	defer conn.Close()

	for _, v := range value {
		_, err := conn.Do("lpush", key, v)
		if err != nil {

			return err
		}
	}

	return nil
}

func (s *RedisPool) LpushCount(key string, number int) error {
	conn := s.pool.Get()
	defer conn.Close()

	for i := 1; i <= number; i++ {
		err := conn.Send("lpush", key, i)
		if err != nil {

			return err
		}
	}
	conn.Flush()

	return nil
}

func (s *RedisPool) LPop(key string) (string, error) {
	conn := s.pool.Get()
	defer conn.Close()

	v, err := conn.Do("lpop", key)
	if err != nil {
		return "", err
	}

	if v == nil {
		return "", nil
	}
	vv := BytesToString(v.([]byte))
	return vv, nil
}

func BytesToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{bh.Data, bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}

func (s *RedisPool) LLen(key string) (int64, error) {
	conn := s.pool.Get()
	defer conn.Close()

	v, err := conn.Do("llen", key)
	if err != nil {

		return 0, err
	}

	if v == nil {
		return 0, nil
	}
	return v.(int64), nil
}

func (s *RedisPool) Close() {
	s.pool.Close()
}

/**
Hincryby方法
*/
func (s *RedisPool) HINCRBY(key, field string) {
	conn := s.pool.Get()
	defer conn.Close()

	conn.Do("HINCRBY", key, field, 1)

}

/**
HGet方法
*/
func (s *RedisPool) HGet(key, field string) (interface{}, error) {
	conn := s.pool.Get()
	defer conn.Close()

	value, err := conn.Do("HGET", key, field)

	return value, err
}

/**
HGetAll方法
*/
func (s *RedisPool) HGetAll(key string) ([][]byte, error) {
	conn := s.pool.Get()
	defer conn.Close()

	inter, err := redis.ByteSlices(conn.Do("HGetAll", key))

	return inter, err
}

/**
Hset方法
*/
func (s *RedisPool) HSet(key string, field string, value string) (interface{}, error) {
	conn := s.pool.Get()
	defer conn.Close()

	inter, err := conn.Do("HSET", key, field, value)

	return inter, err
}

/**
Hset方法
*/
func (s *RedisPool) HSetByte(key string, field string, value []byte) (interface{}, error) {
	conn := s.pool.Get()
	defer conn.Close()

	inter, err := conn.Do("HSET", key, field, value)

	return inter, err
}

/**
SADD方法
*/
func (s *RedisPool) SAdd(args []interface{}) (interface{}, error) {
	conn := s.pool.Get()
	defer conn.Close()

	inter, err := conn.Do("SADD", args...)
	return inter, err
}

/**
Scard方法
*/
func (s *RedisPool) SCard(key string) (interface{}, error) {
	conn := s.pool.Get()
	defer conn.Close()

	inter, err := conn.Do("SCARD", key)
	return inter, err
}

/**
Spop方法
*/
func (s *RedisPool) SPop(key string) (interface{}, error) {
	conn := s.pool.Get()
	defer conn.Close()

	inter, err := conn.Do("SPOP", key)
	vv := BytesToString(inter.([]byte))
	return vv, err
}
