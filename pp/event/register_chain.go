package event

// Author j
import (
	"context"
	"fmt"
	"github.com/qsnetwork/qsds/framework/client/cf"
	"github.com/qsnetwork/qsds/framework/spbf"
	"github.com/qsnetwork/qsds/msg/header"
	"github.com/qsnetwork/qsds/msg/protos"
	"github.com/qsnetwork/qsds/pp/client"
	"github.com/qsnetwork/qsds/pp/serv"
	"github.com/qsnetwork/qsds/pp/setting"
	"github.com/qsnetwork/qsds/utils"
)

// RegisterChain
func RegisterChain(toSP bool) {
	if toSP {
		SendMessageToSPServer(reqRegisterData(toSP), header.ReqRegister)
		utils.Log("SendMessage(conn, req, header.ReqRegister) to SP")
	} else {
		sendMessage(client.PPConn, reqRegisterData(toSP), header.ReqRegister)
		utils.Log("SendMessage(conn, req, header.ReqRegister) to PP")
	}

}

// ReqRegisterChain if get this, must be PP
func ReqRegisterChain(ctx context.Context, conn spbf.WriteCloser) {
	utils.Log("PP get ReqRegisterChain")
	var target protos.ReqRegister
	if unmarshalData(ctx, &target) {
		// store register P wallet address
		serv.RegisterPeerMap.Store(target.Address.WalletAddress, spbf.NetIDFromContext(ctx))
		transferSendMessageToSPServer(reqRegisterDataTR(&target))

		// IPProt := strings.Split(target.Address.NetworkAddress, ":")
		// ip := ""
		// port := ""
		// if len(IPProt) > 1 {
		// 	ip = IPProt[0]
		// 	port = IPProt[1]
		// }
		// if ip == "127.0.0.1" {
		//
		// 	utils.DebugLog("user didn't config network address")
		// 	utils.DebugLog("target", target)
		// 	req := target
		// 	req.Address = &protos.PPBaseInfo{
		// 		WalletAddress:  target.Address.WalletAddress,
		// 		NetworkAddress: conn.(*spbf.ServerConn).GetIP() + ":" + port,
		// 	}
		// 	utils.DebugLog("req", req)
		// 	SendMessageToSPServer(&req, header.ReqRegister)
		// } else {
		// 	// transfer to SP
		// 	transferSendMessageToSPServer(reqRegisterDataTR(&target))
		// }
	}
}

// RspRegisterChain  PP -> SP, SP -> PP, PP -> P
func RspRegisterChain(ctx context.Context, conn spbf.WriteCloser) {
	utils.Log("get RspRegisterChain", conn)
	var target protos.RspRegister
	if unmarshalData(ctx, &target) {

		utils.Log("target.RspRegister", target.WalletAddress)
		if target.WalletAddress == setting.WalletAddress {
			utils.Log("get RspRegisterChain ", target.Result.State, target.Result.Msg)
			if target.Result.State == protos.ResultState_RES_SUCCESS {
				fmt.Println("login successfully", target.Result.Msg)
				setting.IsLoad = true
				utils.DebugLog("@@@@@@@@@@@@@@@@@@@@@@@@@@@@", conn.(*cf.ClientConn).GetName())
				setting.IsPP = target.IsPP
				if !setting.IsPP {
					reportDHInfoToPP()
				}
				if setting.IsAuto {
					if setting.IsPP {
						StartMining()
					}
				}
			} else {
				setting.WalletAddress = ""
				fmt.Println("login failed", target.Result.Msg)
			}

		} else {
			utils.Log("transfer RspRegisterChain to: ", target.WalletAddress)
			transferSendMessageToClient(target.WalletAddress, spbf.MessageFromContext(ctx))
		}
	}
}

// RspMining RspMining
func RspMining(ctx context.Context, conn spbf.WriteCloser) {
	utils.DebugLog("get RspMining", conn)
	var target protos.RspMining
	if unmarshalData(ctx, &target) {
		if target.Result.State == protos.ResultState_RES_SUCCESS {
			fmt.Println("start mining")
			if serv.GetPPServer() == nil {
				go serv.StartListenServer(setting.Config.Port)
			}
			setting.IsSatrtMining = true
			if client.SPConn == nil {
				client.SPConn = client.NewClient(setting.Config.SPNetAddress, setting.IsPP)
				RegisterChain(true)
			}
		} else {
			utils.Log(target.Result.Msg)
		}
	}
}

// StartMining
func StartMining() {
	if setting.CheckLogin() {
		if setting.IsPP {
			utils.DebugLog("StartMining")
			SendMessageToSPServer(reqMiningData(), header.ReqMining)
		} else {
			fmt.Println("register as miner first")
		}
	}
}
