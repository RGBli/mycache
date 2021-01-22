package mycache

/*ByteView 只有一个数据成员，`bytes []byte`
* bytes 将会存储真实的缓存值
* 选择 byte 类型是为了能够支持任意的数据类型的存储，例如字符串、图片等*/
type ByteView struct {
	bytes []byte
}

// 实现 Len() 方法
func (view ByteView) Len() int {
	return len(view.bytes)
}

// 获取 ByteView 的 byte 数组的拷贝，防止外部非法修改
func (view ByteView) ByteSlice() []byte {
	dst := make([]byte, len(view.bytes))
	src := view.bytes
	copy(dst, src)
	return dst
}

// 将 ByteView 的 byte 数组转为字符串
func (view ByteView) ToString() string {
	return string(view.bytes)
}
