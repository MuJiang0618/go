package cache

import (
	"errors"
	"fmt"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// 测试通过
func TestA(t *testing.T) {
	g := NewGroup("lk", 2 << 10, GetterFunc(func(key string) ([]byte, error) {
		v, ok := db[key]
		if ok {
			return []byte(v), nil
		}
		return []byte{}, errors.New("load from local failed!")
	}))

	/* 可以看到日志:
	key: Tom ---- gocache missed
	key: Tom ---- load from local successfully~
	key: Tom ---- gocache hited~
	*/
	_, _ = g.Get("Tom")
	_, _ = g.Get("Tom")
}

func TestB(t *testing.T) {
	g := NewGroup("lk", int64(len("Tom"))+ int64(len("630")), GetterFunc(func(key string) ([]byte, error) {
		v, ok := db[key]
		if ok {
			return []byte(v), nil
		}
		return []byte{}, errors.New("load from local failed!")
	}))

	g.Get("Tom")
	g.Get("Sam")
	g.Get("Tom")
}

func TestC(t *testing.T) {
	g := NewGroup("lk", int64(len("Tom"))+ int64(len("630")), GetterFunc(func(key string) ([]byte, error) {
		v, ok := db[key]
		if ok {
			return []byte(v), nil
		}
		return []byte{}, errors.New("load from local failed!")
	}))

	// 参数中设定cache的容量为一个键值对 "Tom" + "630"
	// 测试通过, 先是加载Tom, 然后Tom被替换为Sam, 然后Sam又被替换为Tom
	g.Get("Tom")
	fmt.Printf("cache size: %d\n", g.cache.size)
	g.Get("Jack")		// 可以看到, Tom被替换为Jack后因为cache的空间超出, Jack键值对被移除了, cache size = 0
	fmt.Printf("cache size: %d\n", g.cache.size)
}