package main

import (
	"./distribution"
	"bytes"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type ClientRemote struct {
	remoteAddr string		// 远程节点的地址
	basePath string			// 远程节点该分布式缓存web服务的根目录 /gocache
	groupName string 		// 远程节点group的名字

	// 处理并发
	mutex sync.Mutex
	reqHandling map[string]*Call
}

type Call struct {
	res []byte		// 存储返回的数据
	ok bool			// 存储请求是否成功的信息
	mutex sync.WaitGroup		// 同步请求同一key的其他协程
}

// 通过http向该client对应的远程节点请求数据
func (clientRemote *ClientRemote) get(key string) ([]byte, bool) {
	// 远程节点收到的请求为: 127.0.0.1:port/basePath?key=.&groupName=.
	// 每个请求挨个检查该 key是否正被处理,这里如果不加锁,即使有一个线程已经标记该 key正被处理
	// 其他请求早已经执行完这些检查代码了,就不会在这里等待结果而是去真正地请求远程节点

	clientRemote.mutex.Lock()
	if call, ok := clientRemote.reqHandling[key]; ok {
		log.Println("waiting......")
		clientRemote.mutex.Unlock()		// 已经检测到该key正被处理, 让锁给其他请求检查 key
		call.mutex.Wait()
		log.Println("终于等到你 ", string(call.res))
		return call.res, call.ok
	}

	time.Sleep(3 * time.Second)		// 可以观察到其他并发请求阻塞等待在 38行代码
	call := &Call{mutex: sync.WaitGroup{}}
	call.mutex.Add(1)
	clientRemote.reqHandling[key] = call
	clientRemote.mutex.Unlock()		// 让其他请求可以检查该 key是否正被处理

	req, err := http.NewRequest(http.MethodGet, "http://" + clientRemote.remoteAddr + clientRemote.basePath, nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	q := req.URL.Query()
	q.Add("key", key)
	q.Add("groupName", clientRemote.groupName)
	req.URL.RawQuery = q.Encode()

	log.Printf("Go %s URL: %s \n", http.MethodGet, req.URL.String())

	client := &http.Client{}
	res, err := client.Do(req)
	if err == nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		value := buf.Bytes()
		call.res = value	// 请求成功返回数据和 true
		call.ok = true

		clientRemote.mutex.Lock()		// 该 key被处理完后删除其被处理状态记录
		delete(clientRemote.reqHandling, key)
		clientRemote.mutex.Unlock()

		//time.Sleep(5 * time.Second)
		call.mutex.Done()

		log.Println("获得了value: ", string(value))

		return value, true
	}

	call.res = nil		// 请求失败返回nil和 false
	call.ok = false

	clientRemote.mutex.Lock()
	delete(clientRemote.reqHandling, key)
	clientRemote.mutex.Unlock()
	call.mutex.Done()

	return nil, false
}

type MasterNode struct {
	groupName string
	slaveNode map[string]*ClientRemote
}

func startSlaveServer(slaveAddr string) {
	//sync.Mutex.Lock()
	router := httprouter.New()
	router.GET("/gocache", distribution.RemoteGet)
	//router.POST("gocache", remotePost)

	log.Printf("启动了从节点 %s", slaveAddr)

	//sync.Mutex.Unlock()
	http.ListenAndServe(slaveAddr, router)
}

func (masterNode *MasterNode) startSlaveNode() {
	for _, clientRemote := range masterNode.slaveNode {
		// clientRemote.remoteAddr == "127.0.0.1:port"
		go startSlaveServer(clientRemote.remoteAddr)		// 启动从节点的web服务
	}
}

func startMasterNode(masterNode *MasterNode) {
	masterNode.startSlaveNode()

	http.Handle("/gocache", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		groupName := req.URL.Query().Get("groupName")

		// clientRemote是一个结构体, 包含对应的从节点的地址
		clientRemote, ok := masterNode.slaveNode[groupName]
		if ok {
			key := req.URL.Query().Get("key")
			value, ok := clientRemote.get(key)
			if ok {
				w.Write(value)
				return
			}
			http.Error(w, "key not found!", http.StatusNotFound)
			return
		}
		http.Error(w, "group not exist!", http.StatusNotFound)
	}))

	http.ListenAndServe(":9000", nil)
}

func startSlaveNode() {
	http.Handle("/hellos", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello slave"))
	}))
	http.ListenAndServe(":9001", nil)
}

func main() {
	slaveNodeMap := map[string]*ClientRemote {
		"group1": &ClientRemote{remoteAddr: "127.0.0.1:9001", basePath: "/gocache", groupName: "group1", reqHandling: make(map[string]*Call), mutex: sync.Mutex{}},
		"group2": &ClientRemote{remoteAddr: "127.0.0.1:9002", basePath: "/gocache", groupName: "group2", reqHandling: make(map[string]*Call), mutex: sync.Mutex{}},
		"group3": &ClientRemote{remoteAddr: "127.0.0.1:9003", basePath: "/gocache", groupName: "group3", reqHandling: make(map[string]*Call), mutex: sync.Mutex{}},
	}

	masterNode := &MasterNode{groupName: "master", slaveNode: slaveNodeMap}
	startMasterNode(masterNode)
}



