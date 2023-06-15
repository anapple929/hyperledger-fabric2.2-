package service

import (
	"fmt"
	"test/sdkInit"

	//"transfer/sdkInit"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
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

/**************************************************************************************************/
type ServiceSetup struct {
	ChaincodeID string
	Client      *channel.Client
}

func InitService(chaincodeID, channelID string, org *sdkInit.OrgInfo, sdk *fabsdk.FabricSDK) (*ServiceSetup, error) {
	handler := &ServiceSetup{
		ChaincodeID: chaincodeID,
	}
	//prepare channel client context using client context
	clientChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org.OrgUser), fabsdk.WithOrg(org.OrgName))
	// Channel client is used to query and execute transactions (Org1 is default org)
	client, err := channel.New(clientChannelContext)
	if err != nil {
		return nil, fmt.Errorf("Failed to create new channel client: %s", err)
	}
	handler.Client = client
	return handler, nil
}
