package frame

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

func TestEncode(t *testing.T) {
	codec := NewMyFrameCodec()
	buf := make([]byte, 0, 128)
	rw := bytes.NewBuffer(buf) // *bytes.Buffer 实现了Write方法，也就是实现了 io.Writer类，可以作为Encode方法的第一个参数

	err := codec.Encode(rw, []byte("hello world"))
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
	}

	// 验证Encode的正确性
	var totalLen int32
	err = binary.Read(rw, binary.BigEndian, &totalLen)
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
	}
	if totalLen != 15 {
		t.Errorf("want 15, actual %d", totalLen)
	}
	left := rw.Bytes()
	if string(left) != "hello world" {
		t.Errorf("want hello world, actual %s", string(left))
	}
}

func TestDecode(t *testing.T) {
	codec := NewMyFrameCodec()
	// 前四位会被解析成 15(0x0000000f)存到totalLen中
	data := []byte{0x0, 0x0, 0x0, 0xf, 'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd'}

	payload, err := codec.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
	}
	if string(payload) != "hello world" {
		t.Errorf("want hello world, actual %s", string(payload))
	}

}

type ReturnErrorWriter struct {
	W  io.Writer // 继承W的所有方法
	Wn int       // 模拟第几次调用Write返回错误
	wc int       // 写操作次数计数
}

func (w *ReturnErrorWriter) Write(p []byte) (int, error) {
	w.wc++
	if w.wc >= w.Wn {
		return 0, errors.New("write error")
	}
	return w.W.Write(p)
}

type ReturnErrorReader struct {
	R  io.Reader
	Rn int // 第几次调用Read返回错误
	rc int // 读操作次数计数
}

func (r *ReturnErrorReader) Read(p []byte) (n int, err error) {
	r.rc++
	if r.rc >= r.Rn {
		return 0, errors.New("read error")
	}
	return r.R.Read(p)
}

func TestEncodeWithWriteFail(t *testing.T) {
	codec := NewMyFrameCodec()
	buf := make([]byte, 0, 128)
	w := bytes.NewBuffer(buf)

	// Encode方法内部调用 binary.Write方法，内部会调用io.Writer.Write方法，会计入 ReturnErrorWriter 的写次数

	// 模拟binary.Write返回错误
	err := codec.Encode(&ReturnErrorWriter{
		W:  w,
		Wn: 1, // 模拟第一次读取失败
	}, []byte("hello"))
	if err == nil {
		t.Errorf("want non-nil, actual nil")
	}

	// 模拟w.Write返回错误
	err = codec.Encode(&ReturnErrorWriter{
		W:  w,
		Wn: 2, // 模拟第二次写失败（写实际数据），第一次写（写长度）成功
	}, []byte("hello"))
	if err == nil {
		t.Errorf("want non-nil, actual nil")
	}
}

func TestDecodeWithReadFail(t *testing.T) {
	codec := NewMyFrameCodec()
	data := []byte{0x0, 0x0, 0x0, 0x9, 'h', 'e', 'l', 'l', 'o'}

	// 模拟binary.Read返回错误
	_, err := codec.Decode(&ReturnErrorReader{
		R:  bytes.NewReader(data),
		Rn: 1,
	})
	if err == nil {
		t.Errorf("want non-nil, actual nil")
	}

	// 模拟io.ReadFull返回错误
	_, err = codec.Decode(&ReturnErrorReader{
		R:  bytes.NewReader(data),
		Rn: 2,
	})
	if err == nil {
		t.Errorf("want non-nil, actual nil")
	}
}
