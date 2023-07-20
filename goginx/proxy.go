package goginx

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

//实现反向代理

// 启动监听服务
func (engine *Engine) startListen() {
	for _, s := range engine.service {
		for _, location := range s.location {
			go location.listen(&engine.mu, &engine.servicesPoll)
		}
	}
}

// 对每个location进行监听
func (location *location) listen(mu *sync.Mutex, servicesPoll *map[string]*location) {
	server := &http.Server{
		Addr: location.root,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch location.locationType {
			case loadBalancing:
				location.hashRing.forward(w, r, mu)
			case fileService:
				location.getFile(w, r, mu)
			}
		}),
	}
	location.httpService = server
	(*servicesPoll)[location.root] = location

	err := server.ListenAndServe()
	if err != nil {
		log.Println("监听", location.root, "错误，错误信息：", err)
	}
}

// 反向代理，将信息转发给后端服务器，再转发回去
func (hashRing hashRing) forward(w http.ResponseWriter, r *http.Request, mu *sync.Mutex) {
	//询问是否正在热重启。如果是则返回503，服务器维护状态码。
	isNotReSet := mu.TryLock()
	if !isNotReSet {
		http.Error(w, "服务重启中，请重试", http.StatusServiceUnavailable)
		return
	} else {
		mu.Unlock()
	}

	// 获取客户端ip
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Println("获取ip错误:", err)
		http.Error(w, "获取ip错误", http.StatusInternalServerError)
		return
	}

	// 获取后端服务器
	serviceIP := hashRing.balancer(ip)
	remote, err := url.Parse("http://" + serviceIP)
	if err != nil {
		log.Println("解析目标服务器地址失败:", err)
		http.Error(w, "解析目标服务器地址失败", http.StatusInternalServerError)
		return
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(w, r)
}
