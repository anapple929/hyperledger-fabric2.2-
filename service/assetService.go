package service

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

func (t *ServiceSetup) SaveUser(userName string, userId string) (string, error) {

	req := channel.Request{ChaincodeID: t.ChaincodeID, Fcn: "userRegister", Args: [][]byte{[]byte(userName), []byte(userId)}}
	respone, err := t.Client.Execute(req)
	if err != nil {
		return "service SaveUser have a error", err
	}

	return string(respone.TransactionID), nil
}

func (t *ServiceSetup) QueryUser(userId string) ([]byte, error) {

	req := channel.Request{ChaincodeID: t.ChaincodeID, Fcn: "queryUser", Args: [][]byte{[]byte(userId)}}
	respone, err := t.Client.Query(req)
	if err != nil {
		return []byte{0x00}, err
	}

	return respone.Payload, nil
}
