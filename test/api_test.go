package test

import (
	"goginx"
	"log"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup

var engine *goginx.Engine

func TestInit(t *testing.T) {
	engine = goginx.Init()
}

func TestStart(t *testing.T) {
	wg.Add(1)
	go engine.Start()
}

func TestReset(t *testing.T) {
	for i := 10; i >= 1; i-- {
		log.Println("reset 还有", i, "s开始")
		time.Sleep(1 * time.Second)
	}
	log.Println("reset测试开始")
	engine.Reset()
	wg.Wait()
}
