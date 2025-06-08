package main

import (
	"37_tcp-server-demo1/frame"
	"37_tcp-server-demo1/packet"
	"fmt"
	"net"
)

// handlePacket 服务端处理客户端发来的 Submit请求，并返回响应
func handlePacket(framePayload []byte) (ackFramePayload []byte, err error) {
	var p packet.Packet // Packet的实现类有 Submit 和 SubmitAck，p的类型一定是二者之一
	p, err = packet.Decode(framePayload)
	if err != nil {
		fmt.Printf("error decoding packet: %v\n", err)
		return
	}

	switch p.(type) { // 类型 switch，根据p的类型进行操作
	case *packet.Submit: // 如果类型为 *Submit，也就是接收到了客户端的请求，只对请求进行处理。
		submit := p.(*packet.Submit) // 类型断言，类似强制类型转换
		fmt.Printf("receive submit: id = %s, payload = %s\n", submit.ID, string(submit.Payload))
		submitAck := &packet.SubmitAck{
			ID:     submit.ID, // 同一次应答保证为同一个ID
			Result: 0,
		}
		ackFramePayload, err = packet.Encode(submitAck)
		if err != nil {
			fmt.Printf("error encoding ack packet: %v\n", err)
			return
		}
		return ackFramePayload, nil
	default:
		return nil, fmt.Errorf("unknwon packet type")
	}
}

// handleConn 处理TCP连接 服务端只需要接受客户端发来的请求，并给予响应
func handleConn(c net.Conn) { // net.Conn接口包含 Read 和 Write函数，实现了io.Reader 和 io.Writer
	defer c.Close()
	frameCodec := frame.NewMyFrameCodec()
	for {
		// 从输入流中读出 framePayLoad 数据（[]byte）
		framePayload, err := frameCodec.Decode(c)
		if err != nil {
			fmt.Printf("error decoding frame: %v\n", err)
			return
		}
		// 解析framePayLoad数据，并得到响应的 ackFramePayload 数据
		ackFramePayload, err := handlePacket(framePayload)
		if err != nil {
			fmt.Printf("error handling packet: %v\n", err)
			return
		}
		// 响应结果传入连接中，并关闭连接(defer c.Close())
		err = frameCodec.Encode(c, ackFramePayload)
		if err != nil {
			fmt.Printf("error encoding ack packet: %v\n", err)
			return
		}
	}
}

func main() {
	l, err := net.Listen("tcp", ":8080") // 服务端监听8080端口
	if err != nil {
		fmt.Printf("Error listening: %s\n", err)
		return
	}
	for {
		conn, err := l.Accept() // 建立 net.Conn 连接
		if err != nil {
			fmt.Printf("Error accepting: %s\n", err)
			break
		}
		go handleConn(conn)
	}
}
