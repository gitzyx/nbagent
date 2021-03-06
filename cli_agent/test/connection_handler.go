package test

import (
	"bytes"
	"context"
	"encoding/binary"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/protocol"
	"git.yayafish.com/nbagent/protocol/demo"
	"git.yayafish.com/nbagent/protocol/node"
	"git.yayafish.com/nbagent/rpc"
	"github.com/golang/protobuf/proto"
	_ "github.com/google/uuid"
)

var (
	RPC_URI_DEMO string = "rpc.demo"
)

type ClientConnectionHandler struct{}

func (this ClientConnectionHandler) OnRead(ptrConnection *network.ServerConnection,
	byteData []byte, ctx context.Context) {
	log.Infof("OnRead local: %v,data: %v", ptrConnection.LocalAddr(), len(byteData))

	var nMsgType uint8
	objReader := bytes.NewBuffer(byteData)
	if szUri, bSuccess := protocol.ReadString(objReader); !bSuccess {
		log.Warningf("read uri fail!")
		return
	} else if anyErr := binary.Read(objReader, binary.LittleEndian, &nMsgType); anyErr != nil {
		log.Warningf("read nMsgType fail!")
		return
	} else if szRequestID, bSuccess := protocol.ReadString(objReader); !bSuccess {
		log.Warningf("read szRequestID fail!")
		return
	} else {
		if nMsgType == protocol.MSG_TYPE_RESPONE {
			ptrServerConnection := network.GetServerConnection(ctx)
			ptrRpcContext, bSuccess := ptrServerConnection.GetRpcContext(szRequestID)
			if bSuccess {
				ptrRpcContext.Msg <- objReader.Bytes()
			} else {
				log.Warningf("can not find response context szRequestID:%v", szRequestID)
			}
			if szUri==rpc.NODE_RPC_RESP{
				objRpcCallResp := node.RpcCallResp{}
				if anyErr := proto.Unmarshal(objReader.Bytes(), &objRpcCallResp); anyErr != nil {
					log.Warningf("proto.Unmarshal error:%v", anyErr)
					return
				}
				log.Infof("RpcCallResp: %+v",objRpcCallResp)
			}
		} else if nMsgType == protocol.MSG_TYPE_REQUEST {
			switch szUri {
			case rpc.NODE_RPC_REQ:
				{
					objRpcCallReq := node.RpcCallReq{}
					if anyErr := proto.Unmarshal(objReader.Bytes(), &objRpcCallReq); anyErr != nil {
						log.Warningf("proto.Unmarshal error:%v", anyErr)
						return
					}
					log.Infof("objRpcCallReq: %+v", objRpcCallReq)

					if objRpcCallReq.URI == RPC_URI_DEMO {
						objTestMsgReq := demo.TestMsgReq{}
						if anyErr := proto.Unmarshal(objRpcCallReq.Data, &objTestMsgReq); anyErr != nil {
							log.Warningf("proto.Unmarshal error:%v", anyErr)
							return
						}
						log.Infof("TestMsgReq: %+v", objTestMsgReq)

						objTestMsgRsp := demo.TestMsgRsp{}
						objTestMsgRsp.TestReply = "Reply: " + objTestMsgReq.TestString
						objRpcCallResp := node.RpcCallResp{}
						objRpcCallResp.Result = node.ResultCode_SUCCESS
						objRpcCallResp.Data, _ = proto.Marshal(&objTestMsgRsp)
						objRpcCallResp.URI = objRpcCallReq.URI
						objRpcCallResp.EntryType = objRpcCallReq.EntryType
						objRpcCallResp.Caller = objRpcCallReq.Caller
						bRet := false
						byteResp, bRet := protocol.NewProtocol(rpc.NODE_RPC_RESP, protocol.MSG_TYPE_RESPONE,
							szRequestID, &objRpcCallResp)
						if !bRet {
							log.Warningf("protocol.NewProtocol error, data: %v", objRpcCallResp)
							return
						}
						bRet = ptrConnection.Write(byteResp)
						if !bRet {
							log.Warningf("Write error, data: %v", objRpcCallResp)
							return
						}
						log.Infof("TestMsgRsp: %+v", objTestMsgRsp)
						log.Infof("objRpcCallResp: %+v", objRpcCallResp)
					}
				}
			}
		}
	}
}

func (this ClientConnectionHandler) OnCloseConnection(ptrConnection *network.ServerConnection) {
	log.Infof("OnCloseConnection local: %v", ptrConnection.LocalAddr())
}