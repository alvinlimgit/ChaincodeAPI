package chaincode

import (
	"encoding/json"
	"fmt"
	"log"
        "bytes"
        "time"
        "strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
        "github.com/hyperledger/fabric-chaincode-go/shim/"
        pb "github.com/hyperledger/fabric-protos-go/peer"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Asset struct {
	ID             string `json:"ID"`
	Color          string `json:"color"`
	Size           int    `json:"size"`
	Owner          string `json:"owner"`
	AppraisedValue int    `json:"appraisedValue"`
}

// Alvin addition. To use getHistoryForKey() METHOD, in ChaincodeStubInterface INTERFACE, in Shim PACKAGE. Function needs to exist in SMART CONTRACT.
// /fabric-samples/asset-transfer-basic/chaincode-go/assetTransfer.go is the executable code, which imports chaincode (this file) & only runs main function
// to initialize. Modular. Import other stuff from other repo, keep main executable code clean.
func (s *SmartContract) AaaBbb(stub shim.ChaincodeStubInterface, args []string) pb.Response {
        if len(args) < 1 {
                return shim.Error("Incorrect number of arguments. Expecting 1")
        }

        assetName := args[0]

        fmt.Printf("- start getHistoryForAsset: %s\n", assetName)

        resultsIterator, err := stub.GetHistoryForKey(assetName)
        if err != nil {
                return shim.Error(err.Error())
        }
        defer resultsIterator.Close()

        // buffer is a JSON array containing historic values for the asset
        var buffer bytes.Buffer
        buffer.WriteString("[")

        bArrayMemberAlreadyWritten := false
        for resultsIterator.HasNext() {
                response, err := resultsIterator.Next()
                if err != nil {
                        return shim.Error(err.Error())
                }

                // Add a comma before array members, suppress it for the first array member
                if bArrayMemberAlreadyWritten == true {
                        buffer.WriteString(",")
                }
                buffer.WriteString("{\"TxId\":")
                buffer.WriteString("\"")
                buffer.WriteString(response.TxId)
                buffer.WriteString("\"")

                buffer.WriteString(", \"Value\":")
                // if it was a delete operation on given key, then we need to set the
                //corresponding value null. Else, we will write the response.Value
                //as-is (as the Value itself a JSON asset)
                if response.IsDelete {
                        buffer.WriteString("null")
                } else {
                        buffer.WriteString(string(response.Value))
                }

                buffer.WriteString(", \"Timestamp\":")
                buffer.WriteString("\"")
                buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

                buffer.WriteString(", \"IsDelete\":")
                buffer.WriteString("\"")
                buffer.WriteString(strconv.FormatBool(response.IsDelete))
                buffer.WriteString("\"")

                buffer.WriteString("}")
                bArrayMemberAlreadyWritten = true
        }
        buffer.WriteString("]")

        fmt.Printf("- getHistoryForAsset returning:\n%s\n", buffer.String())

        return shim.Success(buffer.Bytes())
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "asset1", Color: "rainbow haha", Size: 5, Owner: "Tomoko", AppraisedValue: 300},
		{ID: "asset2", Color: "red", Size: 5, Owner: "Brad", AppraisedValue: 400},
		{ID: "asset3", Color: "green", Size: 10, Owner: "Jin Soo", AppraisedValue: 500},
		{ID: "asset4", Color: "yellow", Size: 10, Owner: "Max", AppraisedValue: 600},
		{ID: "asset5", Color: "black", Size: 15, Owner: "Adriana", AppraisedValue: 700},
		{ID: "asset6", Color: "white", Size: 15, Owner: "Michel", AppraisedValue: 800},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferAsset updates the owner field of asset with given id in world state.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	asset.Owner = newOwner
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
