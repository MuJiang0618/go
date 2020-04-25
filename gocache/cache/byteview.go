package cache

// 将值作为字节数组存储, 这样值就可以支持多种类型如图片
type ByteView struct {
	B []byte // 存放真正的值
}

// 实现Value接口
func (v ByteView) Len() int {
	return len(v.B)
}

// 获取值的拷贝
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.B)
}

// 将值转换为string便于观察
func (v ByteView) String() string {
	return string(v.B)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}