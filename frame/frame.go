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
	err := binary.Write(w, binary.BigEndian, &totalLen)
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
	err := binary.Read(r, binary.BigEndian, &totalLen)
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
