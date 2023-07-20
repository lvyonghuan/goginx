package goginx

import (
	"log"
	"net/http"
	"os"
	"sync"
)

//提供文件服务

func (location *location) getFile(w http.ResponseWriter, r *http.Request, mu *sync.Mutex) {
	//询问是否正在热重启。如果是则返回503，服务器维护状态码。
	isNotReSet := mu.TryLock()
	if !isNotReSet {
		http.Error(w, "服务重启中，请重试", http.StatusServiceUnavailable)
		return
	} else {
		mu.Unlock()
	}

	file, err := os.ReadFile(location.fileRoot)
	if err != nil {
		log.Println("文件查找错误：", err)
		http.Error(w, "文件查找错误", http.StatusInternalServerError)
		return
	}
	contentType := http.DetectContentType(file)
	w.Header().Set("Content-Type", contentType)

	_, err = w.Write(file)
	if err != nil {
		log.Println("写入响应错误：", err)
		http.Error(w, "写入响应错误", http.StatusInternalServerError)
		return
	}
}
