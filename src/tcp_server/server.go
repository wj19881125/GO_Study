package main

import (
	"fmt"
	"net"
)

func main03() {
	var tcpListener, err = net.Listen("tcp", ":1125") //在刚定义好的地址上进监听请求。
	if err != nil {
		fmt.Println("监听出错：", err)
		return
	}
	defer tcpListener.Close()
	fmt.Println("正在等待连接...")
	var conn, err2 = tcpListener.Accept() //接受连接。
	if err2 != nil {
		fmt.Println("接受连接失败：", err2)
		return
	}
	var remoteAddr = conn.RemoteAddr() //获取连接到的对像的IP地址。
	fmt.Println("接受到一个连接：", remoteAddr)
	fmt.Println("正在读取消息...")
	buff := make([]byte, 1024)
	for {
		//var bys, _ = ioutil.ReadAll(conn) //读取对方发来的内容。
		n, err3 := conn.Read(buff)
		if err3 != nil {
			fmt.Println("接受连接失败：", err3)
			return
		}
		fmt.Println("接收到客户端的消息：", string(buff[:n]))
		conn.Write([]byte("hello, Nice to meet you, my name is raiing medical company")) //尝试发送消息。
	}
	conn.Close()
}
