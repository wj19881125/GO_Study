package main

import (
	"fmt"
	"net"
	"strconv"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:1125")
	if err != nil {
		fmt.Println("连接服务器出错, ", err)
		return
	}
	defer conn.Close()
	ch := make(chan int)
	firstName, lastName := "hello", " raiing"
	count := 1
	go func() {
		buff := make([]byte, 1024)
		for {
			fmt.Println(<-ch)
			// 读取数据
			n, err1 := conn.Read(buff)
			if err1 != nil {
				fmt.Println("读取内容出错.", err1)
				break
			}
			fmt.Print("读取内容为，", string(buff[:n]))
		}
	}()
	for {
		fmt.Println(firstName + lastName)
		count ++
		ch <- count
		conn.Write([]byte(firstName + lastName + strconv.Itoa(count)))
		//time.Sleep(time.Second * 1)
	}

}
