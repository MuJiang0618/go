package cache

import (
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("lk")
	if b, _ := f.Get("lk"); !reflect.DeepEqual(expect, b) {
		t.Fatalf("失败啦!")
	}
}