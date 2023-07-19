package goginx

// Init 初始化服务，需要提供哈希环上每个真实节点对应的虚拟节点个数
func Init(replicas int) *Engine {
	engine := createEngine()
	engine.replicas = replicas
	readConfig(&engine)
	return &engine
}

// Start 启动服务
func (engine *Engine) Start() {
	engine.startListen()
}
