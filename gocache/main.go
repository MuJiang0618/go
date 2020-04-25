package main

import (
	"./distribution"
	"bytes"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
)

type ClientRemote struct {
	remoteAddr string		// 远程节点的地址
	basePath string			// 远程节点该分布式缓存web服务的根目录 /gocache
	groupName string 		// 远程节点group的名字
}

func (clientRemote *ClientRemote) get(key string) ([]byte, bool) {
	// 远程节点收到的请求为: 127.0.0.1:port/basePath?key=
	//http.Get(clientRemote.remoteAddr + clientRemote.basePath + "?" + "")
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

		return value, true
	}

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
		"group1": &ClientRemote{remoteAddr: "127.0.0.1:9001", basePath: "/gocache", groupName: "group1"},
		"group2": &ClientRemote{remoteAddr: "127.0.0.1:9002", basePath: "/gocache", groupName: "group2"},
		"group3": &ClientRemote{remoteAddr: "127.0.0.1:9003", basePath: "/gocache", groupName: "group3"},
	}

	masterNode := &MasterNode{groupName: "master", slaveNode: slaveNodeMap}
	startMasterNode(masterNode)
}