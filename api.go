package goginx

// Start 初始化服务，需要提供哈希环上每个真实节点对应的虚拟节点个数
func Start(replicas int) *Engine {
	engine := createEngine()
	engine.replicas = replicas
	readConfig(&engine)
	return &engine
}
