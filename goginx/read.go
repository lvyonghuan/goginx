package goginx

//读取配置文件

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// 配置文件结构
type config struct {
	service  []service
	upstream map[string]*upstream //一个后端服务器池名对应多个后端服务器
}

// upstream结构
type upstream struct {
	addr     []string  //服务器地址
	hashRing *hashRing //upstream对应的哈希环
	replicas int       //每个虚拟节点对应的真实节点数量
}

// service结构
type service struct {
	port        string //定义监听的代理服务器端口号。一个端口号绑定一个service。
	httpService *http.Server
	location    []*location
	hashValue   uint64 //location哈希值的和。
}

// location结构
type location struct {
	locationType int
	root         string //根路径，会附加在service结构的根路径上
	upstream     string //使用的后端服务器池名
	fileRoot     string //fileRoot，文件路径，和root是两个东西了
	hashValue    uint32 //location哈希值。用于验证location是否发生变化。
}

// 读取配置文件
func readConfig(engine *Engine) {
	var cfg config
	cfg.upstream = make(map[string]*upstream)
	file, err := os.Open("./config/config.cfg")
	if err != nil {
		log.Fatalf("open config file failed: %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	//定义当前的type
	const (
		serviceType  = 1
		upstreamType = 2
		locationType = 3
		endType      = 0
	)
	var nowType = 0
	var serviceStruct service
	var locationStruct location
	var upstreamName string
	for scanner.Scan() {

		line := scanner.Text()
		if isSkip(line) {
			continue
		}

		//检测关键字
		switch line {
		case "[server]":
			nowType = serviceType
			continue
		case "[upstream]":
			nowType = upstreamType
			continue
		case "[location]":
			nowType = locationType
			continue
		case "[end]":
			switch nowType {
			case serviceType:
				cfg.service = append(cfg.service, serviceStruct)
				// 重置结构体
				serviceStruct = service{}
				nowType = endType
			case locationType:
				//复制一份
				newLocation := locationStruct
				serviceStruct.location = append(serviceStruct.location, &newLocation)
				locationStruct = location{}
				nowType = serviceType
			case upstreamType:
				upstreamName = ""
				nowType = endType

			}
			continue
		}
		//检查目前字段所处区块
		switch nowType {
		//处理service区块
		case serviceType:
			s := strings.Split(line, "=")
			switch s[0] {
			case "port":
				serviceStruct.port = s[1]
			}
		case upstreamType:
			s := strings.Split(line, "=")
			switch s[0] {
			case "name":
				upstreamName = s[1]
				cfg.upstream[upstreamName] = &upstream{}
			case "replicas":
				replicas, err := strconv.Atoi(s[1])
				if err != nil {
					log.Fatalf("replicas 字段设置错误：%v", err)
				}
				cfg.upstream[upstreamName].replicas = replicas
			default:
				cfg.upstream[upstreamName].addr = append(cfg.upstream[upstreamName].addr, s[0])
			}
		case locationType:
			s := strings.Split(line, "=")
			switch s[0] {
			case "type":
				typeNum, err := strconv.Atoi(s[1])
				if err != nil {
					log.Fatalf("location type字段设置错误：%v", err)
				}
				locationStruct.locationType = typeNum
			case "root":
				locationStruct.root = s[1]
			case "upstream":
				locationStruct.upstream = s[1]
			case "file_root":
				locationStruct.fileRoot = s[1]
			}
		}
	}
	engine.writeEngine(cfg)
}

// 跳过检测。跳过换行符等。
func isSkip(line string) bool {
	if line == "" || line == " " || line == "\n" || line == "\r" || line == "\t" || line[0] == '#' {
		return true
	}
	return false
}
