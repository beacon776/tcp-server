package packet

import (
	"bytes"
	"strings"
	"testing"
)

// Encode: struct -> []byte
// Decode: []byte -> struct
func TestSubmit_Encode(t *testing.T) {
	submit := NewSubmit("12345678", []byte("hello world"))
	encode, err := submit.Encode()
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
		return
	}
	if !bytes.Equal(encode, []byte("12345678hello world")) {
		t.Errorf("want %x, actual %x", []byte("12345678hello world"), encode)
		return
	}
}

func TestSubmit_Encode_Error(t *testing.T) {
	submit := NewSubmit("12345", []byte("hello world"))
	_, err := submit.Encode()
	if err == nil {
		t.Errorf("want error, actual nil")
		return
	}
}

func TestSubmitAck_Encode(t *testing.T) {
	submitAck := NewSubmitAck("12345678", 0)
	encode, err := submitAck.Encode()
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
		return
	}
	if !bytes.Equal(encode, append([]byte("12345678"), 0)) {
		t.Errorf("want %x, actual %x", append([]byte("12345678"), 0), encode)
		return
	}
}

func TestSubmitAck_Encode_Error(t *testing.T) {
	submitAck := NewSubmitAck("12345", 0)
	_, err := submitAck.Encode()
	if err == nil {
		t.Errorf("want error, actual nil")
		return
	}
}

func TestSubmit_Decode(t *testing.T) {
	submitData := []byte("12345678hello world")
	emptySubmit := NewSubmitWithoutParam()
	err := emptySubmit.Decode(submitData)
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
		return
	}
	if emptySubmit.ID != "12345678" {
		t.Errorf("want %s, actual %s", "12345678", emptySubmit.ID)
		return
	}
	if !bytes.Equal(emptySubmit.Payload, []byte("hello world")) {
		t.Errorf("want %s, actual %s", "hello world", emptySubmit.Payload)
		return
	}

}

func TestSubmit_Decode_Error(t *testing.T) {
	invalidData := []byte("12345")
	emptySubmit := NewSubmitWithoutParam()
	err := emptySubmit.Decode(invalidData)
	if err == nil {
		t.Errorf("want packetBody too short, got nil")
		return
	}
}

func TestSubmitAck_Decode(t *testing.T) {
	submitAckData := append([]byte("12345678"), 0)
	emptySubmitAck := NewSubmitAckWithoutParam()
	err := emptySubmitAck.Decode(submitAckData)
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
		return
	}
	if emptySubmitAck.ID != "12345678" {
		t.Errorf("want %s, actual %s", "12345678", emptySubmitAck.ID)
		return
	}

	if emptySubmitAck.Result != 0 {
		t.Errorf("want %x, actual %x", 0, emptySubmitAck.Result)
		return
	}
}

func TestSubmitAck_Decode_Error(t *testing.T) {
	invalidData := append([]byte("12345"), 0)
	emptySubmitAck := NewSubmitAckWithoutParam()
	err := emptySubmitAck.Decode(invalidData)
	if err == nil {
		t.Errorf("want ID must be exactly 8 bytes, got nil")
		return
	}
}

func TestEncode(t *testing.T) {
	submit := NewSubmit("12345678", []byte("hello world"))
	encode, err := Encode(submit)
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
		return
	}
	expected := append([]byte{CommandSubmit}, []byte("12345678hello world")...)
	if !bytes.Equal(encode, expected) {
		t.Errorf("want %x, actual %x", expected, encode)
		return
	}

	submitAck := NewSubmitAck("12345678", 0)
	encode, err = Encode(submitAck)
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
		return
	}
	expected = append([]byte{CommandSubmitAck}, append([]byte("12345678"), 0)...)
	if !bytes.Equal(encode, expected) {
		t.Errorf("want %x, actual %x", expected, encode)
		return
	}

}

func TestEncode_Error(t *testing.T) {
	submit := NewSubmit("12345", []byte("hello world"))
	_, err := Encode(submit)
	if err == nil {
		t.Errorf("want error, actual nil")
		return
	}

	submitAck := NewSubmitAck("12345", 0)
	_, err = Encode(submitAck)
	if err == nil {
		t.Errorf("want error, actual nil")
		return
	}
}

func TestDecode(t *testing.T) {
	packet1 := append([]byte{CommandSubmit}, []byte("12345678hello world")...)
	decode, err := Decode(packet1)
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
		return
	}
	// 注意这里检验 decode 的正确性！
	submit, ok := decode.(*Submit)
	if !ok {
		t.Errorf("want *Submit, actual %x", decode)
		return
	}
	if submit.ID != "12345678" {
		t.Errorf("want %s, actual %s", "12345678", submit.ID)
		return
	}
	if !bytes.Equal(submit.Payload, []byte("hello world")) {
		t.Errorf("want %x, actual %x", []byte("hello world"), submit.Payload)
		return
	}

	packet2 := append([]byte{CommandSubmitAck}, append([]byte("12345678"), 0)...)
	decode2, err := Decode(packet2)
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
		return
	}
	submitAck, ok := decode2.(*SubmitAck)
	if !ok {
		t.Errorf("want *SubmitAck, actual %x", decode2)
		return
	}
	if submitAck.ID != "12345678" {
		t.Errorf("want %s, actual %s", "12345678", submitAck.ID)
	}
	if submitAck.Result != 0 {
		t.Errorf("want %x, actual %x", 0, submitAck.Result)
		return
	}

}

func TestDecode_Error(t *testing.T) {
	// case 1: 空 packet
	_, err := Decode([]byte{})
	if err == nil {
		t.Errorf("want error for empty packet, got nil")
	}

	// case 2: 未知 commandId
	packet1 := append([]byte{0xFF}, []byte("abcdefgh")...)
	_, err = Decode(packet1)
	if err == nil || !strings.Contains(err.Error(), "unknown commandID") {
		t.Errorf("want unknown commandID error, got %v", err)
	}

	// case 3: Submit packet body 太短
	packet2 := append([]byte{CommandSubmit}, []byte("12345")...)
	_, err = Decode(packet2)
	if err == nil || !strings.Contains(err.Error(), "packetBody too short") {
		t.Errorf("want Submit body too short error, got %v", err)
	}

	// case 4: SubmitAck body 不足 9 字节（8 ID + 1 result）
	packet3 := append([]byte{CommandSubmitAck}, []byte("1234567")...)
	_, err = Decode(packet3)
	if err == nil || !strings.Contains(err.Error(), "packetBody too short") {
		t.Errorf("want SubmitAck body too short error, got %v", err)
	}
}
