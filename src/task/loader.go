package task

import (
	"config"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"sync"
)

var TaskList map[string]*taskItem
var configMd5 [16]byte
var localIp string

func Init() {
	TaskList = make(map[string]*taskItem)

	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				localIp = ipNet.IP.String()
			}
		}
	}
}

/**
 * 运行
 */
func Run() {
	fileData := readFile()
	if !checkMd5(fileData) {
		load(fileData)
	}

	wg := sync.WaitGroup{}
	for _, taskItem := range TaskList {
		if taskItem.checkCond() {
			wg.Add(1)
			go taskItem.exec(&wg)
		}
	}

	wg.Wait()
}

/**
 * 读取配置文件
 */
func readFile() []byte {
	tasksFileName := config.GetConfig("tasks")

	fileHandler, err := os.Open(tasksFileName)

	if err != nil {
		fmt.Println("can't find the tasks file!")
		os.Exit(1)
	}
	fileData, _ := ioutil.ReadAll(fileHandler)

	return fileData
}

/**
 * 校验配置文件MD5
 */
func checkMd5(fileData []byte) bool {
	sum := md5.Sum(fileData)
	if len(configMd5) > 0 && sum != configMd5 {
		return false
	}

	return true
}

/**
 * 加载配置
 */
func load(fileData []byte) {
	taskConfigs := make([]attr, 0)
	err := json.Unmarshal(fileData, &taskConfigs)
	if err != nil {
		fmt.Println("task config parse error: " + err.Error())
		os.Exit(1)
	}

	for _, attr := range taskConfigs {
		if TaskList[attr.Command] != nil {
			if reflect.DeepEqual(TaskList[attr.Command].attr, attr) {
				continue
			} else {
				TaskList[attr.Command].clearTask()
			}
		}

		task, err := attr.buildTask()
		if task == nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}

		TaskList[task.attr.Command] = task
	}
}
