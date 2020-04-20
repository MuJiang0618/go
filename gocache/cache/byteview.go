package cache

// 将值作为字节数组存储, 这样值就可以支持多种类型如图片
type ByteView struct {
	b []byte    // 存放真正的值
}

// 实现Value接口
func (v ByteView) Len() int {
	return len(v.b)
}

// 获取值的拷贝
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 将值转换为string便于观察
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

//// 回调函数的接口
//type Getter interface {
//	Get(key string) ([]byte, error)
//}
//
//// 回调函数, 当缓存未命中时, 通过该函数从配置的数据源获取数据
//type GetterFunc func(key string) ([]byte, error)
//
//// Get implements Getter interface function
//func (f GetterFunc) Get(key string) ([]byte, error) {
//	return f(key)
//}
