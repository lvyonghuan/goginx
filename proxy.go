package goginx

import (
	"io"
	"log"
	"net"
	"strconv"
	"sync"
)

var wg sync.WaitGroup //要退出程序的时候直接用一个新的waitGroup罩上去实现归零

//实现反向代理

// 启动监听服务
func (engine *Engine) startListen() {
	const (
		loadBalancing = 1
		fileService   = 2
	)
	for _, s := range engine.service {
		for _, location := range s.location {
			switch location.locationType {
			//如果location类型是负载均衡
			case loadBalancing:
				wg.Add(1)
				go location.listen("127.0.0.1:" + strconv.Itoa(s.listen) + s.root + location.root)
			case fileService:
				//TODO 文件服务
			}
		}
	}
	wg.Wait()
}

// 对每个location进行监听
func (location *location) listen(root string) {
	listener, err := net.Listen("tcp", root)
	if err != nil {
		log.Println("监听", root, "错误，错误信息：", err)
		return
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("监听", root, "出现问题:", err)
			continue
		}
		go location.hashRing.forward(conn)
	}
}

// 反向代理，将信息转发给后端服务器，再转发回去
func (hashRing hashRing) forward(conn net.Conn) {
	defer conn.Close()
	//先获取客户端ip
	var ip string
	clientAddr := conn.RemoteAddr()
	if tcpAddr, ok := clientAddr.(*net.TCPAddr); ok {
		ip = string(tcpAddr.IP)
	} else {
		log.Println("获取ip错误")
		return //一定要获取到ip才能提供负载均衡
	}
	//获取后端服务器
	serviceIP := hashRing.balancer(ip)
	serviceConn, err := net.Dial("tcp", serviceIP)
	if err != nil {
		log.Println("反向代理问题，连接服务器失败。错误信息：", err)
		return
	}
	defer serviceConn.Close()
	_, err = io.Copy(serviceConn, conn)
	if err != nil {
		log.Println("反向代理问题，数据转发到服务器失败，错误信息：", err)
		return
	}
	_, err = io.Copy(conn, serviceConn)
	if err != nil {
		log.Println("反向代理失败，数据转发到客户端失败，错误信息：", err)
	}
}
