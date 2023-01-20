package transport

// package main

import (
	"net"
	"reflect"

	"github.com/sirupsen/logrus"
)

type RockUdp struct {
	Port      int
	RecvBytes uint64
	handler   any
}

// UDP Server端
func (udp *RockUdp) RockUdpServe() (err error) {
	listen, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: udp.Port,
	})
	if err != nil {
		logrus.Fatalf("udp listen failed. port:%d", udp.Port)
		return
	}
	defer listen.Close()

	var data [1024]byte
	for {
		n, fromAddr, err := listen.ReadFromUDP(data[:]) // 接收数据
		if err != nil {
			logrus.Errorf("read udp failed, err: %v", err)
			continue
		}
		// fmt.Println("data:%v fromAddr:%v count:%v\n", string(data[:n]), fromAddr, n)
		// _, err = listen.WriteToUDP(data[:n], fromAddr) // 发送数据
		// if err != nil {
		// 	fmt.Println("Write to udp failed, err: ", err)
		// 	continue
		// }
		intput := make([]reflect.Value, 3)
		intput[0] = reflect.ValueOf(fromAddr)
		intput[1] = reflect.ValueOf(data[:n])
		intput[2] = reflect.ValueOf(n)
		reflect.ValueOf(udp.handler).MethodByName("RockOnMsg").Call(intput)
		// if err = udp.handle(fromAddr, data[:n], n); err != nil {
		// logrus.Errorf("process data error, err:%v", err)
		// }
	}
}
