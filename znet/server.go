package znet

import (
	"fmt"
	"net"

	"github.com/dokidokikoi/my-zinx/utils"
	"github.com/dokidokikoi/my-zinx/ziface"
)

// IServer 的接口实现
type Server struct {
	// 服务器名
	Name string
	// tcp4 or other
	IPVersion string
	// 服务绑定的 IP 地址
	IP string
	// 服务绑定的端口
	Port int
	// 当前 Server 的消息管理模块，用于绑定 MsgID 和对应的处理方法
	msgHandler ziface.IMsgHandler
	// 当前 Server 的连接管理器
	ConnMgr ziface.IConnManager

	onConnStart func(conn ziface.IConnection)
	onConnStop  func(conn ziface.IConnection)

	packet ziface.IPacket
}

func (s *Server) Start() {
	fmt.Printf("[START] Server listener at IP: %s, Port %d, is starting\n", s.IP, s.Port)

	// 开启一个 go 去做服务器的 listener 业务
	go func() {
		// 启动 worker 工作池机制
		s.msgHandler.StartWorkerPool()
		// 1.获取一个 TCP 的 Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr err: ", err)
			return
		}

		// 2.监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}

		// 监听成功
		fmt.Println("start Zinx server", s.Name, " suc, now listening...")

		// TODO: server.go 应该有一个自动生成 id 的方法
		var cid uint32 = 0

		// 3.启动 server 网络连接业务
		for {
			// 3.1 阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}

			// 3.2 设置服务器最大连接控制，
			// 如果超过最大连，则关闭新的连接
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				conn.Close()
				continue
			}

			// 3.3 处理该新连接请求的业务方法，
			// 此时 handler 和 conn 应该是绑定的
			dealConn := NewConnection(s, conn, cid, s.msgHandler)
			cid++

			// 3.4 启动当前连接的处理业务
			go dealConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server, name", s.Name)

	// 将需要清理的连接信息或者其他信息一并停止或者清理
	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	s.Start()

	// TODO: Server.Serve() 如果在启动服务的时候还要处理其他事情，
	// 则可以在这里添加

	// 阻塞，否则主 Go 退出，listener 的 go 将退出
	select {}
}

func (s *Server) AddRouter(msgID uint32, router ziface.IRouter) {
	s.msgHandler.AddRouter(msgID, router)
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

func (s *Server) SetOnConnStart(hookFunc func(ziface.IConnection)) {
	s.onConnStart = hookFunc
}

func (s *Server) SetOnConnStop(hookFunc func(ziface.IConnection)) {
	s.onConnStop = hookFunc
}

func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.onConnStart != nil {
		fmt.Println("----> CallOnConnStart....")
		s.onConnStart(conn)
	}
}

func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.onConnStop != nil {
		fmt.Println("----> CallOnConnStop....")
		s.onConnStop(conn)
	}
}

func (s *Server) Packet() ziface.IPacket {
	return s.packet
}

func NewServer(opts ...Option) ziface.IServer {
	// 先初始化全局配置文件
	utils.GlobalObject.Reload()
	s := &Server{
		Name:       utils.GlobalObject.Name,
		IPVersion:  "tcp4",
		IP:         utils.GlobalObject.Host,
		Port:       utils.GlobalObject.TcpPort,
		msgHandler: NewMsgHandler(),
		ConnMgr:    NewConnManager(),
		packet:     NewDataPack(),
	}

	for _, opt := range opts {
		opt(s)
	}
	return s
}
