package cache

import "encoding/json"

type CacheCodec interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
}

var gCodec CacheCodec

type jsonCodec struct{}

func (c jsonCodec) Encode(val interface{}) ([]byte, error) {
	return json.Marshal(val)
}

func (c jsonCodec) Decode(data []byte, val interface{}) error {
	return json.Unmarshal(data, val)
}

func InitJsonCode() {
	gCodec = jsonCodec{}
}
