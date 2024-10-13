package znet

import (
	"errors"
	"fmt"
	"github.com/dokidokikoi/my-zinx/ziface"
	"sync"
)

type ConnManager struct {
	// 管理的连接信息
	connection map[uint32]ziface.IConnection
	// 读写连接的读写锁
	connLock sync.RWMutex
}

func (cm *ConnManager) Len() int {
	cm.connLock.RLock()
	length := len(cm.connection)
	cm.connLock.RUnlock()
	return length
}

func (cm *ConnManager) Add(conn ziface.IConnection) {
	// 保护共享资源， map 加写锁
	cm.connLock.Lock()
	// 将连接添加到 map 中
	cm.connection[conn.GetConnID()] = conn
	cm.connLock.Unlock()

	fmt.Printf("connection add to ConnManager successfully: conn num=%d\n", cm.Len())
}

// 移除连接，但并未停止连接的业务处理
func (cm *ConnManager) Remove(conn ziface.IConnection) {
	// 保护共享资源， map 加写锁
	cm.connLock.Lock()
	delete(cm.connection, conn.GetConnID())
	cm.connLock.Unlock()

	fmt.Printf("connection Remove ConnID=%d successfully: conn num=%d\n", conn.GetConnID(), cm.Len())
}

func (cm *ConnManager) Get(connID uint32) (ziface.IConnection, error) {
	// 保护共享资源， map 加读锁
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()

	if conn, ok := cm.connection[connID]; ok {
		return conn, nil
	}
	return nil, errors.New("connection not found")
}

func (cm *ConnManager) ClearConn() {
	// 保护共享资源， map 加写锁
	cm.connLock.Lock()

	// 停止并删除所有连接信息
	for connID, conn := range cm.connection {
		conn.Stop()
		delete(cm.connection, connID)
	}
	cm.connLock.Unlock()

	fmt.Printf("Clear All Connection successfully: conn num=%d\n", cm.Len())
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connection: make(map[uint32]ziface.IConnection),
	}
}
