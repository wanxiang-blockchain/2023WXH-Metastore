package cache

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"sync"
)

type CacheType uint8

const (
	MAP CacheType = iota
	REDIS
)

type Cache interface {
	Get([]byte) ([]byte, error)
	Put([]byte, []byte) error
	Delete([]byte) error
}

var gCache Cache

type MapCache struct {
	m map[string][]byte
	sync.RWMutex
}

func (c *MapCache) Get(key []byte) ([]byte, error) {
	c.RLock()
	defer c.RUnlock()
	return c.m[string(key)], nil
}
func (c *MapCache) Put(key []byte, value []byte) error {
	c.Lock()
	defer c.Unlock()
	c.m[string(key)] = value
	return nil
}
func (c *MapCache) Delete(key []byte) error {
	c.Lock()
	defer c.Unlock()
	delete(c.m, string(key))
	return nil
}
func NewMapCache() Cache {
	return &MapCache{
		m: map[string][]byte{},
	}
}

func InitGlobalCache(cacheType CacheType) error {
	switch cacheType {
	case MAP:
		gCache = NewMapCache()
		return nil
	}
	return fmt.Errorf("unknown cache type: %d", cacheType)
}

type withCacheConfig struct {
	cache Cache
	codec CacheCodec
}

type withCacheOpt func(*withCacheConfig)

func WithCustomizedCache(c Cache) withCacheOpt {
	return func(wcc *withCacheConfig) {
		wcc.cache = c
	}
}

func WithCustomizedCodec(c CacheCodec) withCacheOpt {
	return func(wcc *withCacheConfig) {
		wcc.codec = c
	}
}

func defaultCacheConfig() *withCacheConfig {
	return &withCacheConfig{
		cache: gCache,
		codec: gCodec,
	}
}

// use cache when calling functions/methods
// if passing argument not function, will return the argument back
// function uses global cache and codec defined in this package
// the args and outputs of the decorated function MUST be exported
func WithCache[T any](f T, opts ...withCacheOpt) T {
	fn := reflect.ValueOf(f)
	for fn.Kind() == reflect.Pointer {
		fn = fn.Elem()
	}
	if fn.Kind() != reflect.Func {
		return f
	}
	cfg := defaultCacheConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	v := reflect.MakeFunc(fn.Type(), func(args []reflect.Value) (results []reflect.Value) {
		funcName := getFunctionName(fn)
		argsStr, err := valuesToByte(args, cfg.codec)
		key := cacheKey(funcName, argsStr)
		data, err := cfg.cache.Get([]byte(key))
		if err != nil {
			return fn.Call(args)
		}
		if data != nil {
			tokens := bytes.Split(data, []byte{0})
			for i, token := range tokens {
				outi := fn.Type().Out(i)
				valI := reflect.New(outi).Interface()
				err := cfg.codec.Decode(token, valI)
				if err != nil {
					return fn.Call(args)
				}
				results = append(results, reflect.ValueOf(valI).Elem())
			}
			return results
		}
		ret := fn.Call(args)
		b, err := valuesToByte(ret, cfg.codec)
		if err == nil {
			cfg.cache.Put(key, b)
		}
		return ret
	})
	return v.Interface().(T)
}

func cacheKey(funcName string, argStr []byte) []byte {
	b := bytes.NewBuffer(nil)
	b.WriteString(funcName)
	b.Write(argStr)
	return b.Bytes()
}

func getFunctionName(v reflect.Value) string {
	return runtime.FuncForPC(v.Pointer()).Name()
}

func valuesToByte(vals []reflect.Value, codec CacheCodec) ([]byte, error) {
	b := bytes.NewBuffer(nil)
	for i, arg := range vals {
		argI := arg.Interface()
		val, err := codec.Encode(argI)
		if err != nil {
			return nil, err
		}
		b.Write(val)
		if len(vals)-1 != i {
			b.WriteByte(0)
		}
	}
	return b.Bytes(), nil
}
