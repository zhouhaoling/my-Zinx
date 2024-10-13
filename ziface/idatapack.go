package ziface

//// 封包数据和拆包数据
//// 直接面向 TCP 连接中的数据流，为传输数据添加头部信息
//// 用于处理 TCP 粘包问题
//type IDataPack interface {
//	// 获取包头长度
//	GetHeadLen() uint32
//	// 封包
//	Pack(msg IMessage) ([]byte, error)
//	// 拆包
//	Unpack([]byte) (IMessage, error)
//}
