package cache

import (
	"log"
	"sync"
)

type Group struct {
	name string
	getter Getter
	cache *Cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// 回调函数的接口
type Getter interface {
	Get(key string) ([]byte, error)
}

// 回调函数, 当缓存未命中时, 通过该函数从配置的数据源获取数据
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// NewGroup create a new instance of Group
func NewGroup(name string, capacity int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		//cache: Cache{capacity: capacity, cache: make(map[string]*list.Element), dequeue: list.New(), size: int64(0)},
		cache: NewCache(capacity),
	}

	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if len(key) == 0 {
		panic("key is required!")
	}

	if value, ok := g.cache.Get(key); ok {
		log.Printf("key: %s ---- gocache hited~\n", key)
		return value.(ByteView), nil
	} else {
		log.Printf("key: %s ---- gocache missed\n", key)
		return g.load(key)			// 首先从本节点磁盘加载数据, 失败则从远程其他节点加载数据
	}
}

func (g *Group) load(key string) (ByteView, error){
	bytes, err := g.loadLocal(key)
	if err == nil {
		return ByteView{b: bytes}, nil
	}
	//else if bytes, err := g.loadDistant(key); err == nil {
	//	return ByteView{b: bytes}, nil
	//}

	return ByteView{}, err
}

func (g *Group) loadLocal(key string) ([]byte, error){
	bytes, err := g.getter.Get(key)
	if  err == nil {
		log.Printf("key: %s ---- load from local successfully~\n", key)
		g.populateCache(key, bytes)
		return bytes, nil
	} else {			// 如果本地加载失败, 从远程节点加载
		log.Printf("key: %s ---- load from local failed!\n", key)
		return []byte{}, err
	}

	//g.populateCache(key, bytes)
	//return ByteView{b: bytes}, nil
}

func (g *Group) populateCache(key string, bytes []byte) {
	g.cache.Add(key, ByteView{b: bytes})
}