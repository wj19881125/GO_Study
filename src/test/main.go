package main

import (
	"fmt"
	"strconv"
	"time"
)

//var a = 12
// const A = 12
//var b = "string"

//var c interface{}
//
//type d interface {
//
//}
//type e int

var arr [10] int

var slice [6]int

var m map[int]string

type KK struct {
	a int
	b int
}

func main02() {

	data := make(chan []string, 10)

	count := 0

	go func() {
		for {
			count++
			data <- []string{"raiing" + strconv.Itoa(count)}
			fmt.Println("子协程", "raiing"+strconv.Itoa(count))
		}
	}()

	for {
		data2, ok := <-data
		if ok {
			fmt.Println("主协程，", data2)
		}
		time.Sleep(time.Second)
	}

	//return
	//inta := 10
	//intb := 3
	////fmt.Println(float64(inta) / float64(intb))
	//s := fmt.Sprintf("%f", float64(inta)/float64(intb))
	//fmt.Println(s)
	//
	//cc := KK{a: 12}
	//fmt.Println(cc)
	//cc.b = 13
	//fmt.Println(cc)
	//dd := make(map[string]*KK, 10)
	//dd["1223"] = &KK{a: 12}
	//fmt.Println(dd)
	//ee := dd["1223"]
	//ee.b = 14
	//for k, v := range dd {
	//	fmt.Println(k, v.a, v.b)
	//}
	//
	//return
	//m = make(map[int]string)
	//m[1] = "1"
	//m[2] = "2"
	//for index := range m {
	//	fmt.Println(index, m[index])
	//}
	//delete(m, 2)
	//data, ok := m[2]
	//if ok {
	//	fmt.Println(ok, data)
	//} else {
	//	fmt.Println(ok)
	//}

	//for i := 0; i < len(arr); i++ {
	//	fmt.Println(arr[i])
	//}
	//for i, data := range arr {
	//	fmt.Printf("索引: %d, 值: %d", i, data)
	//	fmt.Printf("\n")
	//}
	//slice = make([]int, 10)
	//slice[8] = 10
	//slice = [6]int{1, 2}
	//numbers := append(slice, 11)
	//fmt.Println("切片长度: " + strconv.Itoa(len(slice)) + "切片容量: " + strconv.Itoa(cap(slice)))
	//fmt.Println("切片长度: " + strconv.Itoa(len(numbers)) + "切片容量: " + strconv.Itoa(cap(numbers)))
	//ch := make(chan int)
	//count := 0
	//go func() {
	//	for {
	//		fmt.Println(<-ch)
	//	}
	//}()
	//
	//for {
	//	count ++
	//	ch <- count
	//	time.Sleep(time.Second * 2)
	//}

	//fmt.Println("hello world!")
	//fmt.Println("你好",a,b)
	//fmt.Println("你好",c)
	//fmt.Println("你好",new(d))
	//e := 100
	//fmt.Println("你好",e)
}
