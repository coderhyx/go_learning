package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

type CustomData struct {
	ID   uint32
	Name [32]byte
}

func main() {
	// 创建服务端 Socket
	addr, _ := net.ResolveUDPAddr("udp", "localhost:8888")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	fmt.Println("服务端启动，等待客户端连接...")

	// 接收客户端数据
	buffer := make([]byte, 36)
	for {
		_, clientAddr, _ := conn.ReadFromUDP(buffer)

		// 解析数据
		var data CustomData
		//小端序
		data.ID = binary.LittleEndian.Uint32(buffer[:4])
		copy(data.Name[:], buffer[4:])

		fmt.Printf("接收到客户端数据：ID=%d, Name=%s\n", data.ID, string(data.Name[:]))

		// 处理数据...

		// 响应客户端
		response := []byte("Hello, client!")
		conn.WriteToUDP(response, clientAddr)
	}
}
