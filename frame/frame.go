package frame

import (
	"encoding/binary"
	"errors"
	"io"
)

type FramePayload []byte

type StreamFrameCodec interface {
	Encode(io.Writer, FramePayload) error   // data -> Frame -> io.Writer
	Decode(io.Reader) (FramePayload, error) // io.Reader -> FramePayload
}

var (
	ErrShortRead  = errors.New("short read")
	ErrShortWrite = errors.New("short write")
)

type myFrameCodec struct{}

func NewMyFrameCodec() StreamFrameCodec {
	return &myFrameCodec{}
}
func (*myFrameCodec) Encode(w io.Writer, framePayload FramePayload) error {
	var f = framePayload
	var totalLen int32 = int32(len(framePayload)) + 4
	err := binary.Write(w, binary.BigEndian, &totalLen) // 将一个 int32 类型的 totalLen 按照大端字节序（高位字节在前，低位字节在后）写入到 io.Writer 中
	/*
		底层逻辑如下：
		// 1. 把 int32 按照大端序编码成字节
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(totalLen))

		// 2. 把字节写入 w（io.Writer）
		_, err := w.Write(b)
	*/
	if err != nil {
		return err
	}
	n, err := w.Write([]byte(f))
	if err != nil {
		return err
	}
	if n != len(framePayload) {
		return ErrShortWrite
	}
	return nil
}

func (*myFrameCodec) Decode(r io.Reader) (FramePayload, error) {
	var totalLen int32
	err := binary.Read(r, binary.BigEndian, &totalLen) // 从r中读取 totalLen 类型（int32）所需的字节数（4字节），并按照指定的大端字节序填充到data中
	if err != nil {
		return nil, err
	}
	buf := make([]byte, totalLen-4)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	if n != int(totalLen-4) {
		return nil, ErrShortRead
	}
	return FramePayload(buf), nil
}
