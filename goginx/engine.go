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
	upstream          map[string]*upstream
	servicesPoll      map[string]*service //现有的服务池
	resetServicesPoll map[string]*service //重启后的服务池
	state             int                 //引擎现在的状态
	wg                sync.WaitGroup
}

func createEngine() *Engine {
	engine := Engine{}
	engine.resetServicesPoll = make(map[string]*service)
	engine.servicesPoll = make(map[string]*service)
	return &engine
}

func (engine *Engine) writeEngine(cfg config) {
	engine.mu.Lock()
	engine.service = cfg.service
	engine.upstream = cfg.upstream

	//处理后端服务器池，建构哈希环
	for _, v := range engine.upstream {
		v.hashRing = &hashRing{}
		v.hashRing.nodes = make(map[int]string)
		v.addNode(engine)
	}

	//处理服务节点
	for i := range engine.service {
		service := &engine.service[i]
		for _, location := range service.location {
			//计算location哈希值，用于一致性比对
			location.hashValue = hash([]byte(strconv.Itoa(location.locationType) + location.root + location.fileRoot + location.upstream))
			service.hashValue += uint64(location.hashValue)
		}
		if engine.state == reset { //reset信息写入reset map
			engine.resetServicesPoll[service.port] = service
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
		//首先确定不存在的，启动服务
		src, ok := engine.servicesPoll[key]
		if !ok {
			engine.wg.Add(1)
			go value.listen(&engine.mu, &engine.servicesPoll, &engine.upstream, &engine.wg)
			continue
		}
		//如果已经存在，则确认哈希value是否是一致的
		if value.hashValue != src.hashValue {
			//如果不等于，则停掉原来的服务，再根据新的重启
			delete(engine.servicesPoll, key)
			err := src.httpService.Close()
			if err != nil {
				log.Println("关闭服务错误：", err)
			}
			engine.wg.Add(1)
			go value.listen(&engine.mu, &engine.servicesPoll, &engine.upstream, &engine.wg)
		}
	}
	//确认已经关掉的服务
	for key, value := range engine.servicesPoll {
		_, ok := engine.resetServicesPoll[key]
		if !ok {
			delete(engine.servicesPoll, key)
			err := value.httpService.Close()
			if err != nil {
				log.Println("关闭服务错误：", err)
			}
		}
	}
	engine.resetServicesPoll = make(map[string]*service) //释放内存
	engine.mu.Unlock()
}

func (engine *Engine) stopEngine() {
	for _, value := range engine.servicesPoll {
		value.httpService.Close()
	}
	log.Println("程序退出")
}
