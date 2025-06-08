package packet

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	CommandConn   = iota + 0x01 // 0x01  连接请求包
	CommandSubmit               // 0x02 消息请求包
)

const (
	CommandConnAck   = iota + 0x80 // 0x81 连接响应包
	CommandSubmitAck               //0x82 消息响应包
)

type Packet interface {
	Decode([]byte) error     // []byte -> struct
	Encode() ([]byte, error) // struct -> []byte
	/* 虽然 Encode 无传入参数，Decode 除了error外无返回参数，但是请注意：方法的执行者（receiver）会作为第一个参数传入函数参数列表，
	相当于有了对应的传参和修改参数。
	*/
}

type Submit struct {
	ID      string // 消息流水号（请求和响应的ID保持一致）
	Payload []byte // 消息的有效载荷
}
type SubmitAck struct { // SubmitAck 是 Submit Acknowledgement 的缩写，表示提交应答
	ID     string // 消息流水号（请求和响应的ID保持一致）
	Result uint8  // 响应状态（0：正常，1：错误）
}

func NewSubmitWithoutParam() *Submit {
	return &Submit{}
}

func NewSubmit(ID string, Payload []byte) *Submit {
	return &Submit{
		ID:      ID,
		Payload: Payload,
	}
}

func NewSubmitAckWithoutParam() *SubmitAck {
	return &SubmitAck{}
}

func NewSubmitAck(ID string, Result uint8) *SubmitAck {
	return &SubmitAck{
		ID:     ID,
		Result: Result,
	}
}

/* 先声明出 Submit 以及 SubmitAck 两种类型的 Encode 和 Decode 方法，
再声明出通用的 Encode 和 Decode函数，根据CommandID字段选择对应的方法
*/

func (p *Submit) Decode(packetBody []byte) error {
	if packetBody == nil {
		return errors.New("packetBody is nil")
	}
	// 各具体类型分别负责检查各自的传参长度(最低情况为8个字节，也就是无Payload）
	if len(packetBody) < 8 {
		return errors.New("packetBody too short")
	}
	p.ID = string(packetBody[:8])
	p.Payload = packetBody[8:]
	return nil
}
func (p *Submit) Encode() ([]byte, error) {
	if len(p.ID) != 8 {
		return nil, errors.New("ID must be exactly 8 bytes")
	}
	return bytes.Join([][]byte{[]byte(p.ID[:8]), p.Payload}, nil), nil
	/*
		func Join(s [][]byte, sep []byte) []byte
		s数组为需要连接的多个切片， sep为分隔符
	*/
}

func (p *SubmitAck) Decode(packetBody []byte) error {
	if packetBody == nil {
		return errors.New("packetBody is nil")
	}
	// 必须要有Result这一位，也就是说，最低的情况为9个字节
	if len(packetBody) < 9 {
		return errors.New("packetBody too short")
	}
	p.ID = string(packetBody[:8])
	p.Result = packetBody[8]
	return nil
}

func (p *SubmitAck) Encode() ([]byte, error) {

	if len(p.ID) < 8 {
		return nil, errors.New("ID too short")
	}
	if p.Result != 0 && p.Result != 1 {
		return nil, errors.New("result must be 0 or 1")
	}

	return bytes.Join([][]byte{[]byte(p.ID[:8]), []byte{p.Result}}, nil), nil
}

func Encode(p Packet) ([]byte, error) {
	if p == nil {
		return nil, errors.New("packet is nil")
	}
	var (
		commandID  uint8
		packetBody []byte
		err        error
	)
	switch t := p.(type) {
	case *Submit:
		commandID = CommandSubmit
		packetBody, err = t.Encode()
		if err != nil {
			return nil, err
		}
	case *SubmitAck:
		commandID = CommandSubmitAck
		packetBody, err = t.Encode()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown packet type [%s]", t)
	}
	return bytes.Join([][]byte{[]byte{commandID}, packetBody}, nil), nil
}

func Decode(packet []byte) (Packet, error) {
	if packet == nil || len(packet) == 0 {
		return nil, errors.New("packet is nil")
	}

	commandId := packet[0]
	packetBody := packet[1:]
	switch commandId {
	case CommandConn:
		return nil, nil
	case CommandConnAck:
		return nil, nil
	case CommandSubmit:
		s := Submit{}
		err := s.Decode(packetBody) // 注意，Decode时修改了s的内容
		if err != nil {
			return nil, err
		}
		return &s, nil
	case CommandSubmitAck:
		s := SubmitAck{}
		err := s.Decode(packetBody) // 注意，Decode时修改了s的内容
		if err != nil {
			return nil, err
		}
		return &s, nil
	default:
		return nil, fmt.Errorf("unknown commandID [%d]", commandId)
	}
}
