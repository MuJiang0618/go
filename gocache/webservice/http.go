package webservice

import (
	"../cache"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const Default_Base_Path string = "/gocache"   // gocache web服务的项目名

/*
httproute中参数:
/gocache/:groupname/:key
这样就定义了两个参数groupname和key, 要在路径中注明, 而不能用?groupname= &key=
 */

func Get(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	// 请求路径为: /gocache/groupname/key
	// 先找到group, 再用该group查找key
	groupAndKey := strings.SplitN(r.URL.Path, "/", -1)
	if len(groupAndKey) < 4 {
		http.Error(w, "both group & key is required!", http.StatusBadRequest)
	}

	groupName := param.ByName("group")
	key := param.ByName("key")
	if(len(key) == 0) {
		http.Error(w, "key is required!", http.StatusBadRequest)
		return
	}

	group, ok := cache.GetGroup(groupName, nil, false)
	if !ok {
		http.Error(w, "no such group: " + groupName, http.StatusNotFound)
		return
	}

	byteView, err := group.Get(key)
	if err != nil {
		http.Error(w, "no such entry!", http.StatusNotFound)
		return
	}

	log.Printf("查询key: %s 成功~", key)
	_, _ = w.Write(byteView.B)
}

func Post(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	// /gocache/groupname/key/
	groupAndKey := strings.SplitN(r.URL.Path, "/", -1)
	if len(groupAndKey) < 5 {
		http.Error(w, "group & key & value required!", http.StatusBadRequest)
	}

	groupName := param.ByName("group")
	key := param.ByName("key")
	value := param.ByName("value")

	if(len(groupName) == 0 || len(key) == 0 || len(value) == 0) {
		http.Error(w, "参数不能为空!", http.StatusBadRequest)
		return
	}

	group, ok := cache.GetGroup(groupName, nil, true)
	if !ok {
		http.Error(w, "no such group!" + groupName, http.StatusNotFound)
		return
	}

	group.Add(key, value)
	log.Printf("添加key: %s 成功~", key)

	w.Write([]byte("添加key成功~"))
}

type httpClient struct {
	baseURL string		// 该客户端对应的服务区的名称也即地址
}

func (h *httpClient) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}