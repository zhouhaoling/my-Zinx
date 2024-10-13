package znet

import (
	"bytes"
	"fmt"
	"github.com/dokidokikoi/my-zinx/ziface"
	"io"
	"testing"
)

func perparData() []byte {
	dp := NewDataPack()
	msg1 := &Message{
		ID:      1,
		DataLen: 5,
		Data:    []byte("hello"),
	}

	sendData1, err := dp.Pack(msg1)
	if err != nil {
		fmt.Println("Client pack msg1 err: ", err)
		return nil
	}

	msg2 := &Message{
		ID:      2,
		DataLen: 7,
		Data:    []byte("world!!"),
	}

	sendData2, err := dp.Pack(msg2)
	if err != nil {
		fmt.Println("Client pack msg2 err: ", err)
		return nil
	}

	return append(sendData1, sendData2...)
}

func handle1Data(conn io.Reader) ziface.IMessage {
	dp := NewDataPack()
	headData := make([]byte, dp.GetHeadLen())
	_, err := io.ReadFull(conn, headData)
	if err != nil {
		fmt.Println("read head error")
		return nil
	}
	msgHead, err := dp.Unpack(headData)
	if err != nil {
		fmt.Println("server unpack error: ", err)
		return nil
	}

	if msgHead.GetDataLen() > 0 {
		msg := msgHead.(*Message)
		msg.Data = make([]byte, msg.GetDataLen())
		_, err := io.ReadFull(conn, msg.Data)
		if err != nil {
			fmt.Println("read data error")
			return nil
		}
	}
	return msgHead
}

func TestPackUnpack(t *testing.T) {
	conn := bytes.NewReader(perparData())
	for {
		msg := handle1Data(conn)
		if msg == nil {
			return
		}
		fmt.Printf("==> Recv Msg: ID=%d, len=%d, data=%s\n", msg.GetMsgID(), msg.GetDataLen(), msg.GetData())
	}
}
