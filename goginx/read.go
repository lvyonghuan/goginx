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
	upstream map[string][]string //一个后端服务器池名对应多个后端服务器
}

// service结构
type service struct {
	listen   int    //定义监听的代理服务器端口号
	root     string //文件根位置
	location []*location
}

// location结构
type location struct {
	replicas     int
	locationType int
	root         string   //根路径，会附加在service结构的根路径上
	upstream     string   //使用的后端服务器池名
	hashRing     hashRing //哈希环
	httpService  *http.Server
	fileRoot     string
}

// 读取配置文件
func readConfig(engine *Engine) {
	var cfg config
	cfg.upstream = make(map[string][]string)
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
			case "listen":
				port, err := strconv.Atoi(s[1])
				if err != nil {
					log.Fatalf("read config file failed: %v", err)
				}
				serviceStruct.listen = port
			case "root":
				serviceStruct.root = s[1]
			}
		case upstreamType:
			s := strings.Split(line, "=")
			switch s[0] {
			case "name":
				upstreamName = s[1]
			default:
				cfg.upstream[upstreamName] = append(cfg.upstream[upstreamName], s[0])
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
			case "replicas":
				replicas, err := strconv.Atoi(s[1])
				if err != nil {
					log.Fatalf("replicas 字段设置错误：%v", err)
				}
				locationStruct.replicas = replicas
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
