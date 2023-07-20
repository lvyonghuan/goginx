package goginx

//引擎控制

import (
	"log"
	"strconv"
	"sync"
)

// location type描述
const (
	loadBalancing = 1
	fileService   = 2
)

// 引擎状态描述
const (
	start = 1
	run   = 2
	reset = 3 //实际有用的好像就这个
)

type Engine struct {
	mu                sync.Mutex //一把锁，用于动态修改引擎
	service           []service
	upstream          map[string][]string
	wg                sync.WaitGroup       //要退出程序的时候直接用一个新的waitGroup罩上去实现归零
	servicesPoll      map[string]*location //现有的服务池
	resetServicesPoll map[string]*location //重启后的服务池
	state             int                  //引擎现在的状态
}

func createEngine() *Engine {
	engine := Engine{}
	engine.resetServicesPoll = make(map[string]*location)
	engine.servicesPoll = make(map[string]*location)
	return &engine
}

func (engine *Engine) writeEngine(cfg config) {
	engine.mu.Lock()
	engine.service = cfg.service
	engine.upstream = cfg.upstream
	for _, service := range engine.service {
		for _, location := range service.location {
			//重写root
			var root string
			switch location.locationType {
			case loadBalancing:
				root = "127.0.0.1:" + strconv.Itoa(service.listen) + service.root + location.root
			case fileService:
				location.fileRoot = location.root
				root = "127.0.0.1:" + strconv.Itoa(service.listen) + service.root
			}
			location.root = root

			//建构哈希环
			location.hashRing.nodes = make(map[int]string)
			location.addNode(engine)
			if engine.state == reset { //reset信息写入reset map
				engine.resetServicesPoll[root] = location
			}
		}
	}
	if engine.state != reset { //如果engine状态等于reset，将在重写完成之后再启动
		engine.mu.Unlock()
	}
}

func (engine *Engine) resetEngine() {
	engine.state = reset
	readConfig(engine)
	for key, value := range engine.resetServicesPoll {
		if _, ok := engine.servicesPoll[key]; !ok {
			engine.wg.Add(1)
			go value.listen(&engine.mu, &engine.servicesPoll)
		}
	}
	for key, value := range engine.servicesPoll {
		if _, ok := engine.resetServicesPoll[key]; !ok {
			err := value.httpService.Close()
			if err != nil {
				log.Println("关闭服务错误：", err)
			}
			engine.wg.Done()
		} else {
			value.hashRing = engine.resetServicesPoll[key].hashRing
		}
	}
	engine.mu.Unlock()
}
