package utils

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dokidokikoi/my-zinx/utils/commandline/args"
	"io/ioutil"
	"os"

	"github.com/dokidokikoi/my-zinx/ziface"
)

// 全局参数
type GlobalObj struct {
	// 当前 Zinx 全局 Server 对象
	TcpServer ziface.IServer
	Host      string
	TcpPort   int
	Name      string
	Version   string

	// 读取数据包的最大值
	MaxPacketSize uint32
	// 当前服务器主机允许的最大连接个数
	MaxConn int
	// 业务工作池的数量
	WorkerPoolSize uint32
	// 业务工作 worker 对应任务队列的最大任务存储数量
	MaxWorkerTaskLen uint32
	MaxMsgChanLen    uint32

	ConfFilePath string

	// 日志所在文件夹
	LogDir string
	// 日志文件名称，默认 ""
	LogFile string
	// 是否关闭 debug 日志级别调试信息，默认打开
	LogDebugClose bool
}

var GlobalObject *GlobalObj

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, nil
}

// 加载用户的配置文件
func (g *GlobalObj) Reload() {
	if confFileExists, _ := PathExists(g.ConfFilePath); !confFileExists {
		fmt.Println("加载默认配置")
		return
	}
	data, err := ioutil.ReadFile(GlobalObject.ConfFilePath)
	if err != nil {
		fmt.Println("加载默认配置")
		return
	}

	// 将 json 数据解析到 struct 中
	err = json.Unmarshal(data, GlobalObject)
	if err != nil {
		panic(err)
	}
}

func init() {
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "."
	}

	// 初始化配置模块 flag
	args.InitConfigFlag("./conf/zinx.json", "配置文件，如果没有设置，则默认为<exeDir>/conf/zinx.json")
	flag.Parse()
	args.FlagHandle()
	// 初始化 GlobalObject 变量，设置一些默认值
	GlobalObject = &GlobalObj{
		Name:             "ZinxServerApp",
		Version:          "v0.10",
		TcpPort:          7777,
		Host:             "0.0.0.0",
		MaxPacketSize:    4096,
		ConfFilePath:     args.Args.ConfigFile,
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024,
		MaxMsgChanLen:    1024,
		MaxConn:          12000,
		LogDir:           pwd + "/log",
		LogFile:          "",
		LogDebugClose:    false,
	}

	// 从配置文件加载用户配置
	GlobalObject.Reload()
}
