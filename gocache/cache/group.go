package cache

import (
	"errors"
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

type String string

func (s String) Len() int {
	return len(s)
}

// 回调函数, 当缓存未命中时, 通过该函数从配置的数据源获取数据
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// line:62 GetGroup()和该函数有锁冲突, 需要解决
func NewGroup(name string, capacity int64, getter Getter) *Group {
	log.Printf("new group created")
	if getter == nil {
		log.Printf("nil Getter! 该group无法从本地加载数据!")
		// panic("nil getter!)
	}
	//mu.Lock()
	//defer mu.Unlock()
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
func GetGroup(name string, getter Getter, createIfNotExist bool) (*Group, bool) {
	mu.RLock()
	g, ok := groups[name]
	if (!ok) {
		if !createIfNotExist {
			return nil, false
		} else {
			g = NewGroup(name, 2 <<10, getter)
			//groups[name] = g
			return g, true
		}
	}
	mu.RUnlock()
	return g, true
}

func (g *Group) Add(key string, value string) {
	g.cache.Add(key, ByteView{B: []byte(value)})
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
		if g.getter == nil {
			return ByteView{}, errors.New("not found")
		}
		return g.load(key)			// 首先从本节点磁盘加载数据, 失败则从远程其他节点加载数据
	}
}

func (g *Group) load(key string) (ByteView, error){
	bytes, err := g.loadLocal(key)
	if err == nil {
		log.Printf("从数据源加载数据成功~")
		return ByteView{B: bytes}, nil
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
	g.cache.Add(key, ByteView{B: bytes})
}