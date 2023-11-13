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
	// 创建客户端 Socket
	serverAddr, _ := net.ResolveUDPAddr("udp", "localhost:8888")
	conn, _ := net.DialUDP("udp", nil, serverAddr)
	defer conn.Close()

	// 封装数据
	data := CustomData{
		ID:   123,
		Name: [32]byte{'H', 'e', 'l', 'l', 'o'},
	}

	// 序列化数据
	buffer := make([]byte, 36)
	binary.LittleEndian.PutUint32(buffer[:4], data.ID)
	copy(buffer[4:], data.Name[:])

	// 发送数据
	conn.Write(buffer)

	// 接收服务端响应
	response := make([]byte, 1024)
	n, _ := conn.Read(response)
	fmt.Println("接收到服务端响应：", string(response[:n]))
}
