package goginx

// Init 初始化服务，需要提供哈希环上每个真实节点对应的虚拟节点个数
func Init() *Engine {
	engine := createEngine()
	readConfig(engine)
	return engine
}

// Start 启动服务
func (engine *Engine) Start() {
	engine.startListen()
	engine.wg.Wait()
}

// Reset 重启动服务，不中断服务。
func (engine *Engine) Reset() {
	engine.resetEngine()
	engine.wg.Wait()
}

func (engine *Engine) Stop() {
	engine.stopEngine()
}
