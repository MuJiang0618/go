package cache

import (
	"testing"
)

type String string

// 实现 Value接口
func (d String) Len() int {
	return len(d)
}

// 测试能否顺利调用Cache.Get(), 测试通过
func TestAdd(t *testing.T) {
	//t.SkipNow()
	Cache := NewCache(2 << 10, )
	Cache.Add("key1", String("12345"))
	if v, ok := Cache.Get("key1"); !ok || string(v.(String)) != "12345" {
		t.Fatalf("cache hit key1=1234 failed")
	}

	if _, ok := Cache.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

// 测试是否在超出内存时会调用RemoveOldest(), 测试通过
func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"

	Cache := NewCache(int64(len(k1)) + int64(len(k2)) + int64(len(v1)) + int64(len(v2)))
	Cache.Add(k1, String(v1))
	Cache.Add(k2, String(v2))
	Cache.Add(k3, String(v3))

	if _, ok := Cache.Get(k1); ok {
		t.Fatalf("替换失败!")
	}
}