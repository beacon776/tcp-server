package main

import (
	"37_tcp-server-demo1/frame"
	"37_tcp-server-demo1/packet"
	"fmt"
	"github.com/lucasepe/codename" // 第三方包 记得 go mod tidy哈
	"net"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	var num int = 5 // 模拟5个子 goroutine
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func(i int) {
			defer wg.Done()
			startClient(i) // 每个 goroutine 执行各自的流程：向客户端发起请求，并处理客户端发出的响应
		}(i + 1)
	}
	wg.Wait()
}

func startClient(i int) {
	quit := make(chan struct{})
	done := make(chan struct{})
	conn, err := net.Dial("tcp", ":8080") // 向8080端口发起请求
	if err != nil {
		fmt.Printf("dial error: %v\n", err)
		return
	}
	defer conn.Close() // 退出前断开连接
	fmt.Printf("[client %d]: dial ok", i)

	// 利用第三方包 codename 随机生成请求的 payload
	rng, err := codename.DefaultRNG()
	if err != nil {
		panic(err)
	}
	frameCodec := frame.NewMyFrameCodec()
	var counter int // 计数，记录ID

	go func() {
		// 这个 goroutine 是在处理服务端给的响应，把它明明为 响应 goroutine
		for {
			select { // 用来控制请求发完后的退出
			case <-quit: // 响应 goroutine 准备退出
				done <- struct{}{} // 响应goroutine 准备退出后，给 主goroutine 的done channel发消息，让主goroutine退出
				return
			default:
			}

			conn.SetReadDeadline(time.Now().Add(time.Second * 5)) // 设置读取数据最大时间为5s
			ackFramePayLoad, err := frameCodec.Decode(conn)       // 服务端已经把响应写入流了，客户端这边读取流中的响应内容
			if err != nil {
				if e, ok := err.(net.Error); ok {
					if e.Timeout() { // 如果5s内没有服务端响应，Decode会报 timeout 错误，则继续让 goroutine 等待下一次服务端响应
						continue
					}
				}
				panic(err)
			}

			p, err := packet.Decode(ackFramePayLoad) // 二级decode，把 ackFramePayload 转成 SubmitAck 对象
			submitAck, ok := p.(*packet.SubmitAck)   // 类型断言，检测 submitAck 的类型是否为 *SubmitAck
			if !ok {
				panic("not submitAck")
			}
			fmt.Printf("[client %d]: the result of submit ack[%s] is %d\n", i, submitAck.ID, submitAck.Result)
		}
	}()

	// 这个goroutine 是向服务端提交请求
	for {
		// 提交请求
		counter++
		id := fmt.Sprintf("%08d", counter)   // Sprintf 是格式化指定指定格式（比如填零，对齐等）字符串的函数，先把整数格式化为%08d.再转为字符串
		payload := codename.Generate(rng, 4) // 随机生成请求的 payload 内容
		s := &packet.Submit{                 // 构建 *Submit 对象
			ID:      id,
			Payload: []byte(payload),
		}

		framePayload, err := packet.Encode(s) // 二级编码，把 s 对象转为 byte[]
		if err != nil {
			panic(err)
		}
		fmt.Printf("[client %d]: send submit id = %s, payload = %s, frame lenth = %d\n",
			i, s.ID, string(s.Payload), len(framePayload)+4)
		err = frameCodec.Encode(conn, framePayload) // 一级编码成内容为framePayload的 []byte数组，并传入流中
		if err != nil {
			panic(err)
		}

		time.Sleep(time.Second * 1)
		if counter >= 10 { // 一个 goroutine 发送十次请求
			quit <- struct{}{} // 配合select case 里的 <-quit，达到退出的目的
			<-done             // done channel 接收到 响应goroutine 的消息后，让主goroutine退出，否则会阻塞在这里
			fmt.Printf("[client %d] exist ok\n", i)
			return
		}
	}
}
