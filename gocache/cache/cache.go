package cache

import (
	"container/list"
	"sync"
)

type Cache struct {
	mutex sync.Mutex
	cache map[string]*list.Element    // 用来根据键查找值
	dequeue *list.List                   // 双向链表, 从队头删除最久未使用元素
	capacity int64						// 单位: 字节, 考虑的是双向链表中键和值的大小之和
	size int64     // 已经
}

type Value interface {
	Len() int
}

type Entry struct {
	key string
	value Value
}

func  NewCache(capacity int64) *Cache {
	return &Cache{
		cache: make(map[string]*list.Element),
		dequeue: list.New(),
		capacity: capacity,
		size: 0,
	}
}

// 键: 字符串, 值: 多种类型. 值的类型须实现len方法用于计算键值对的大小,防止超出缓存拥有的空间
func (c *Cache) Add(key string, value Value) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 键已存在, 覆盖值
	if ele, ok := c.cache[key]; ok {
		c.dequeue.MoveToFront(ele)			// 更新到最近访问
		kv := ele.Value.(*Entry)
		c.size += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.dequeue.PushFront(&Entry{key, value})
		c.cache[key] = ele
		c.size += int64(len(key)) + int64(value.Len())
	}

	// 添加节点后判断是否超出空间限制, 清除最久未使用的节点
	for c.capacity > 0 && c.capacity < c.size {
		c.RemoveOldest()
	}
}

// 当用户删除某个键值对时调用
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(key) == 0 {
		panic("key is required!")
	}

	if ele, ok := c.cache[key]; ok {
		c.dequeue.Remove(ele)
		kv := ele.Value.(*Entry)
		delete(c.cache, kv.key)
		c.size -= int64(len(kv.key)) + int64(kv.value.Len())
	}
}

func (c *Cache) RemoveOldest() {
	if len(c.cache) == 0 {
		return
	}

	ele := c.dequeue.Back()
	kv := ele.Value.(*Entry)
	delete(c.cache, kv.key)
	c.dequeue.Remove(ele)
	c.size -= int64(len(kv.key)) + int64(kv.value.Len())
}



func (c *Cache) Get(key string) (value Value, ok bool) {
	if len(key) == 0 {
		panic("key is required!")
	}

	if ele, ok := c.cache[key]; ok {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		c.dequeue.MoveToFront(ele)			// 将最近访问的节点移到队头
		kv := ele.Value.(*Entry)			// 用了反射
		return kv.value, true
	}

	return nil, false
}

// 返回缓存的键值对数目
func (c *Cache) Len() int64 {
	return int64(c.dequeue.Len())
}