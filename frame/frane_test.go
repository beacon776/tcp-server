package frame

import (
	"bytes"
	"encoding/binary"
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
