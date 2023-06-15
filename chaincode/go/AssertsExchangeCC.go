package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type AssertsExchangeCC struct {
}

const (
	originOwner = "originOwnerPlaceholder"
)

//User
type User struct {
	Name   string   `json:"name"`
	Id     string   `json:"id"`
	Assets []string `json:"assets"` //AssetId
}

//Assert
type Asset struct {
	Name string `json:"name"`
	Id   string `json:"id"`
	//Metadata map[string]string `json:"metadata"` //special attribute
	Metadata string `json:"metadata"`
}

//AssertExchange in history
type AssetHistory struct {
	AssertId       string `json:"asset_id"`
	OriginOwnerId  string `json:"origin_owner_id"`  //history owner
	CurrentOwnerId string `json:"current_owner_id"` //now owner
}

/************************************************************************************************************************/
func constructUserKey(userId string) string {
	return fmt.Sprintf("user_%s", userId)
}
func constructAssetKey(assetId string) string {
	return fmt.Sprintf("asset_%s", assetId)
}

/*
func constructAssetHistoryKey(originUserId,assetId,currentUserId string) string{
	return fmt.Sprintf("history_%s_%s_%s",originUserId,assetId,currentUserId)
}*/
/************************************************************************************************************************/
//User account opening
func userRegister(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//step 1: check   how many args
	if len(args) != 2 {
		return shim.Error("not enough args,should 2,from userRegister")
	}

	//step 2: check arg is correct?
	name := args[0]
	id := args[1]
	if name == "" || id == "" {
		return shim.Error("invalid args")
	}

	//step 3: check arg is exist?   1.should exist  2. not should exist    it should read database,so add stub in args
	if userBytes, err := stub.GetState(constructUserKey(id)); err == nil && len(userBytes) != 0 {
		return shim.Error("user already exist")
	}
	//step 4: write state
	user := &User{
		Name:   name,
		Id:     id,
		Assets: make([]string, 0),
	}
	//serialize
	userBytes, err := json.Marshal(user)
	if err != nil {
		return shim.Error(fmt.Sprintf("Marshal error:%s", err))
	}
	if err := stub.PutState(constructUserKey(id), userBytes); err != nil {
		return shim.Error(fmt.Sprintf("put user error:%s", err))
	}

	return shim.Success(nil)
}

//User account closing
func userDestory(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//step 1: check   how many args
	if len(args) != 1 {
		return shim.Error("not enough args,should 1,from userDestory")
	}
	//step 2: check arg is correct?
	id := args[0]
	if id == "" {
		return shim.Error("invalid args")
	}
	//step 3: check arg is exist?   1.should exist  2. not should exist    it should read database,so add stub in args
	userBytes, err := stub.GetState(constructUserKey(id))
	if err != nil || len(userBytes) == 0 {
		return shim.Error("user not exist")
	}
	//step 4: write state
	if err := stub.DelState(constructUserKey(id)); err != nil {
		return shim.Error(fmt.Sprintf("delete user error:%s", err))
	}

	//delete user assets
	user := new(User)
	if err := json.Unmarshal(userBytes, user); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal user error:%s", err))
	}
	//todo bug
	for _, assetid := range user.Assets {
		if err := stub.DelState(constructAssetKey(assetid)); err != nil {
			return shim.Error(fmt.Sprintf("delete asset error %s", err))
		}
	}

	return shim.Success(nil)
}

//Asset Enroll
func assetEnroll(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//step 1: check   how many args
	if len(args) != 4 {
		return shim.Error("not enough args,should 4,from assetEnroll")
	}

	//step 2: check arg is correct?
	assetName := args[0]
	assetId := args[1]
	metadata := args[2]
	ownerId := args[3]
	if assetName == "" || assetId == "" || ownerId == "" {
		return shim.Error("invalid args")
	}

	//step 3: check arg is exist?   1.should exist  2. not should exist    it should read database,so add stub in args
	userBytes, err := stub.GetState(constructUserKey(ownerId))
	if err != nil || len(userBytes) == 0 {
		return shim.Error("user not exist")
	}

	if assetBytes, err := stub.GetState(constructAssetKey(assetId)); err == nil && len(assetBytes) != 0 {
		return shim.Error("asset already exist")
	}

	//step 4: write state
	//4.1 write Asset
	asset := &Asset{
		Name:     assetName,
		Id:       assetId,
		Metadata: metadata,
	}
	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal asset error:%s", err))
	}
	if err := stub.PutState(constructAssetKey(assetId), assetBytes); err != nil {
		return shim.Error(fmt.Sprintf("save asset error:%s", err))
	}

	//4.2 update User
	user := new(User)
	if err := json.Unmarshal(userBytes, user); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal user error:%s", err))
	}
	user.Assets = append(user.Assets, assetId)

	userBytes, err = json.Marshal(user)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal user error:%s", err))
	}
	if err := stub.PutState(constructUserKey(user.Id), userBytes); err != nil {
		return shim.Error(fmt.Sprintf("save user error:%s", err))
	}
	//4.3 write Asset history
	history := &AssetHistory{
		AssertId:       assetId,
		OriginOwnerId:  originOwner,
		CurrentOwnerId: ownerId,
	}
	historyBytes, err := json.Marshal(history)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal assert history error:%s", err))
	}

	historyKey, err := stub.CreateCompositeKey("history", []string{
		assetId,
		originOwner,
		ownerId,
	})

	if err != nil {
		return shim.Error(fmt.Sprintf("create key error:%s", err))
	}

	if err := stub.PutState(historyKey, historyBytes); err != nil {
		return shim.Error(fmt.Sprintf("save assert history error:%s", err))
	}
	return shim.Success(nil)
}

//Asset Exchange
func assetExchange(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//step 1: check   how many args
	if len(args) != 3 {
		return shim.Error("not enough args,should 3,from assetExchange")
	}
	//step 2: check arg is correct?
	ownerId := args[0]
	assetId := args[1]
	currentOwnerId := args[2]
	if ownerId == "" || assetId == "" || currentOwnerId == "" {
		return shim.Error("invalid args")
	}

	//step 3: check arg is exist?   1.should exist  2. not should exist    it should read database,so add stub in args
	originOwnerBytes, err := stub.GetState(constructUserKey(ownerId))
	if err != nil || len(originOwnerBytes) == 0 {
		return shim.Error("user not exist")
	}

	currentOwnerBytes, err := stub.GetState(constructUserKey(currentOwnerId))
	if err != nil || len(currentOwnerBytes) == 0 {
		return shim.Error("currentUser not exist")
	}

	assetBytes, err := stub.GetState(constructAssetKey(assetId))
	if err != nil || len(assetBytes) == 0 {
		return shim.Error("asset not exist")
	}

	//check originOwner have asset
	originOwner := new(User)
	if err := json.Unmarshal(originOwnerBytes, originOwner); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal user error:%s", err))
	}

	aidExist := false
	for _, aid := range originOwner.Assets {
		if aid == assetId {
			aidExist = true
			break
		}
	}
	if !aidExist {
		return shim.Error("asset owner not match:%s")
	}
	//4.write state
	//4.1 origin owner delete asset
	assetIds := make([]string, 0)
	for _, aid := range originOwner.Assets {
		if aid == assetId {
			continue
		}
		assetIds = append(assetIds, aid)
	}
	originOwner.Assets = assetIds
	originOwnerBytes, err = json.Marshal(originOwner)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal user error:%s", err))
	}
	if err := stub.PutState(constructUserKey(ownerId), originOwnerBytes); err != nil {
		return shim.Error(fmt.Sprintf("put user error:%s", err))
	}
	//4.2 new owner add asset
	currentOwner := new(User)
	if err := json.Unmarshal(currentOwnerBytes, currentOwner); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal user error:%s", err))
	}
	currentOwner.Assets = append(currentOwner.Assets, assetId)

	currentOwnerBytes, err = json.Marshal(currentOwner)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal user error:%s", err))
	}
	if err := stub.PutState(constructUserKey(currentOwnerId), currentOwnerBytes); err != nil {
		return shim.Error(fmt.Sprintf("put user error:%s", err))
	}
	//4.3 asset history add
	history := &AssetHistory{
		AssertId:       assetId,
		OriginOwnerId:  ownerId,
		CurrentOwnerId: currentOwnerId,
	}
	historyBytes, err := json.Marshal(history)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal assert history error:%s", err))
	}

	historyKey, err := stub.CreateCompositeKey("history", []string{
		assetId,
		ownerId,
		currentOwnerId,
	})

	if err != nil {
		return shim.Error(fmt.Sprintf("create key error:%s", err))
	}

	if err := stub.PutState(historyKey, historyBytes); err != nil {
		return shim.Error(fmt.Sprintf("save assert history error:%s", err))
	}

	return shim.Success(nil)
}

//query User
func queryUser(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//step 1: check   how many args
	if len(args) != 1 {
		return shim.Error("not enough args,should 1,from queryUser")
	}

	//step 2: check arg is correct?
	ownerId := args[0]
	if ownerId == "" {
		return shim.Error("invalid args")
	}

	//step 3: check arg is exist?   1.should exist  2. not should exist    it should read database,so add stub in args
	userBytes, err := stub.GetState(constructUserKey(ownerId))
	if err != nil || len(userBytes) == 0 {
		return shim.Error("user not exist")
	}

	return shim.Success(userBytes)
}

//query Asset
func queryAsset(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//step 1: check   how many args
	if len(args) != 1 {
		return shim.Error("not enough args,should 1,from queryAsset")
	}

	//step 2: check arg is correct?
	assetId := args[0]
	if assetId == "" {
		return shim.Error("invalid args")
	}

	//step 3: check arg is exist?   1.should exist  2. not should exist    it should read database,so add stub in args
	assetBytes, err := stub.GetState(constructAssetKey(assetId))
	if err != nil || len(assetBytes) == 0 {
		return shim.Error("asset not exist")
	}

	return shim.Success(assetBytes)
}

//query Asset History
func queryAssetHistory(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//step 1: check   how many args
	if len(args) != 1 && len(args) != 2 {
		return shim.Error("not enough args,should 1 or 2,from queryAssetHistory")
	}

	//step 2: check arg is correct?
	assetId := args[0]

	if assetId == "" {
		return shim.Error("invalid args")
	}
	queryType := "all"
	if len(args) == 2 {
		queryType = args[1]
	}
	if queryType != "all" && queryType != "enroll" && queryType != "exchange" {
		return shim.Error("not this type queryType")
	}

	//step 3: check arg is exist?   1.should exist  2. not should exist    it should read database,so add stub in args
	assetBytes, err := stub.GetState(constructAssetKey(assetId))
	if err != nil || len(assetBytes) == 0 {
		return shim.Error("asset not exist")
	}
	//step 4: search
	keys := make([]string, 0)
	keys = append(keys, assetId)
	switch queryType {
	case "enroll":
		keys = append(keys, originOwner)
	case "exchange", "all":

	default:
		return shim.Error(fmt.Sprintf("unSupport Type %s", queryType))
	}
	result, err := stub.GetStateByPartialCompositeKey("history", keys)
	if err != nil {
		return shim.Error(fmt.Sprintf("query history error: %s", err))
	}
	defer result.Close()
	histories := make([]*AssetHistory, 0)
	for result.HasNext() {
		historyVal, err := result.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("query error:%s", err))
		}

		history := new(AssetHistory)
		if err := json.Unmarshal(historyVal.GetValue(), history); err != nil {
			return shim.Error(fmt.Sprintf("unmarshal error:%s", err))
		}

		if queryType == "enroll" && history.OriginOwnerId == originOwner {
			continue
		}

		histories = append(histories, history)
	}
	historiesBytes, err := json.Marshal(histories)
	if err != nil {
		return shim.Error(fmt.Sprintf("unmarshal err %s", err))
	}
	return shim.Success(historiesBytes)
}

/**************************************************************************************************************/

func (c *AssertsExchangeCC) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}
func (c *AssertsExchangeCC) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "userRegister":
		return userRegister(stub, args)
	case "userDestory":
		return userDestory(stub, args)
	case "assetEnroll":
		return assetEnroll(stub, args)
	case "assetExchange":
		return assetExchange(stub, args)
	case "queryUser":
		return queryUser(stub, args)
	case "queryAsset":
		return queryAsset(stub, args)
	case "queryAssetHistory":
		return queryAssetHistory(stub, args)
	default:
		return shim.Error(fmt.Sprintf("unsupported function %s", funcName))
	}
	if len(args) != 2 {
		return shim.Error("Incorrect arguments")
	}
	return shim.Success(nil)
}
func main() {
	err := shim.Start(new(AssertsExchangeCC))
	if err != nil {
		fmt.Printf("main run error: %s", err)
	}
}
