package distribution

import (
	"../cache"
	"github.com/julienschmidt/httprouter"
	"hash/crc32"
	"net/http"
	"sort"
	"strconv"
)

/*
设置一个主节点, 用户和该主节点交互, 无法感知其他节点, 通过主节点调用其他节点
根据查询参数key找到虚拟节点, 再找到真实节点, 再拿到该真实节点对应的方法(因为不同节点http请求路径不同
不同真实节点对应的方法的不同点就在于路径
这个路径
 */

// Hash maps bytes to uint32
type HashFunc func(data []byte) uint32

// Map constains all hashed keys
type Map struct {
	hashFunc HashFunc
	replicas int		// 虚拟节点个数
	keys []int 			// Sorted 保存所有虚拟节点的hash值
	hashMap  map[int]string		// 键是虚拟节点的哈希值，值是真实节点的名称
}

// New creates a Map instance
func NewMap(replicas int, fn HashFunc) *Map {
	m := &Map{
		replicas: replicas,
		hashFunc:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE		// 默认方法
	}
	return m
}

// Add adds some keys to the hash.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {	// 创建虚拟节点
			hash := int(m.hashFunc([]byte(strconv.Itoa(i) + key)))		// 计算虚拟节点hash值
			m.keys = append(m.keys, hash)		//
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

var (
	groups = make(map[string]*cache.Group)
)

func init() {
	groups["group1"] = cache.NewGroup("group1", 2 << 10, nil)
	groups["group1"].Add("lk", "lh")

	groups["group3"] = cache.NewGroup("group3", 2 << 10, nil)
	groups["group3"].Add("lh", "lk")
}
// 根据查询的key返回真实节点的名字
//func (m *Map) Get(key string) string {
//	if len(m.keys) == 0 {
//		return ""
//	}
//
//	hash := int(m.hashFunc([]byte(key)))
//	// Binary search for appropriate replica.
//	idx := sort.Search(len(m.keys), func(i int) bool {		// 找到key对应的虚拟节点的位置
//		return m.keys[i] >= hash
//	})
//
//	return m.hashMap[m.keys[idx%len(m.keys)]]		// return真实节点的名字
//}
//
//func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
//	peers := geecache.NewHTTPPool(addr)		// 第一个参数是主节点地址
//	peers.Set(addrs...)		// 注册从节点
//	gee.RegisterPeers(peers)
//	log.Println("geecache is running at", addr)
//	log.Fatal(http.ListenAndServe(addr[7:], peers))
//}

func RemoteGet(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	groupName := r.URL.Query().Get("groupName")
	group, ok := groups[groupName]
	if ok {
		key := r.URL.Query().Get("key")
		byteView, err := group.Get(key)
		if err == nil {
			w.Write(byteView.B)
			return
		}
		http.Error(w, "key not found!", http.StatusNotFound)
		return
	}

	http.Error(w, "no such group!", http.StatusNotFound)
	return
}