package znet

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/dokidokikoi/my-zinx/utils"
	"github.com/dokidokikoi/my-zinx/ziface"
)

type DataPack struct{}

func (dp *DataPack) GetHeadLen() uint32 {
	// ID uint32(4字节) + DataLen uint32(4字节)
	return 8
}

func (dp *DataPack) Pack(msg ziface.IMessage) ([]byte, error) {
	// 创建一个存放 byte 的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	// 写 dataLen
	err := binary.Write(dataBuff, binary.LittleEndian, msg.GetDataLen())
	if err != nil {
		return nil, err
	}

	// 写 msgID
	err = binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgID())
	if err != nil {
		return nil, err
	}

	// 写 data 数据
	err = binary.Write(dataBuff, binary.LittleEndian, msg.GetData())
	if err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

func (dp *DataPack) Unpack(binaryData []byte) (ziface.IMessage, error) {
	// 创建一个输入二进制数据的 ioReader
	dataBuff := bytes.NewBuffer(binaryData)

	// 只解压 head 的信息，得到 dataLen 和 msgID
	msg := &Message{}

	// 读 dataLen
	err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen)
	if err != nil {
		return nil, err
	}

	// 读 msgID
	err = binary.Read(dataBuff, binary.LittleEndian, &msg.ID)
	if err != nil {
		return nil, err
	}

	// 判断 dataLen 的长度是否超出允许的范围
	if utils.GlobalObject.MaxPacketSize > 0 &&
		msg.DataLen > utils.GlobalObject.MaxPacketSize {
		return nil, errors.New("Too Large msg data received")
	}

	// 这里只需把 head 的数据拆包出来即可
	// 然后通过 head 的长度从 conn 读取一次数据
	return msg, nil
}

func NewDataPack() ziface.IPacket {
	return &DataPack{}
}
