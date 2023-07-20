package main

import (
	"bufio"
	"goginx/goginx"

	"log"
	"os"
	"strings"
	"sync"
)

func main() {
	const (
		start = "start"
		reset = "reset"
		stop  = "stop"
	)
	reader := bufio.NewReader(os.Stdin)
	var engine *goginx.Engine // 声明一个 Engine 变量
	var mu sync.Mutex         // 互斥锁，用于保护 engine 变量的访问

	for {
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)

		switch command {
		case start:
			mu.Lock()
			if engine == nil {
				engine = goginx.Init()
				go engine.Start()
			} else {
				log.Println("goginx: engine already started")
			}
			mu.Unlock()
		case reset:
			mu.Lock()
			if engine != nil {
				engine.Reset()
			} else {
				log.Println("goginx: engine not started yet")
			}
			mu.Unlock()
		case stop:
			mu.Lock()
			if engine != nil {
				engine.Stop()
				engine = nil // 释放引用，方便重新启动
			} else {
				log.Println("goginx: engine not started yet")
			}
			mu.Unlock()
		default:
			log.Println("goginx: error command")
		}
	}
}
