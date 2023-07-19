package goginx

//引擎控制

import "sync"

type Engine struct {
	mu       sync.Mutex //一把锁，用于动态修改引擎
	service  []service
	upstream map[string][]string
	replicas int //每个节点对应的虚拟节点数量，手动设置（哈希环参数）
}

func createEngine() Engine {
	return Engine{}
}

func (engine *Engine) writeEngine(cfg config) {
	engine.mu.Lock() //锁住engine，以便热修改进行
	engine.service = cfg.service
	engine.upstream = cfg.upstream
	//建构哈希环
	for _, service := range engine.service {
		for _, location := range service.location {
			location.hashRing.nodes = make(map[int]string)
			location.addNode(engine)
		}
	}
}
