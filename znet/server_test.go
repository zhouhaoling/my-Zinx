package znet

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func ClientTest() {
	fmt.Println("Client Test ... Start")

	// 3s 后发起测试请求，给服务器端开启服务的计会
	time.Sleep(3 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("Client start err, exit!")
		return
	}

	for {
		_, err := conn.Write([]byte("Hello Zinx"))
		if err != nil {
			fmt.Println("write error", err)
			return
		}

		buf := make([]byte, 512)
		cnt, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read buf error")
			return
		}

		fmt.Printf("Server call back: %s, cnt = %d\n", buf, cnt)

		time.Sleep(time.Second)
	}
}

func TestServer(t *testing.T) {
	// 1. 创建一个 Server 句柄
	s := NewServer()

	go ClientTest()

	// 2. 开启服务
	s.Serve()
}
