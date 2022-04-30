package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Asset struct {
	ID    string `json:"ID"`
	Name  string `json:"name"`
	Count uint64    `json:"count"`
	OwnerId        string `json:"ownerId"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface, id string, count uint64) error {
	asset := Asset{id + "-0", id, count, "0"}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(asset.ID, assetJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	return nil
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

func (s *SmartContract) ReadHistory(ctx contractapi.TransactionContextInterface, id string) ([]*Asset, error) {
	stateQuery, err := ctx.GetStub().GetHistoryForKey(id)
	println(stateQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if stateQuery == nil {
		return nil, fmt.Errorf("the history %s does not exist")
	}

	var assets []*Asset
	if err != nil {
		return nil, err
	}
	for stateQuery.HasNext() {
		queryResponse, err := stateQuery.Next()
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
	stateQuery.Close()
	return assets, nil
}

// DeleteAsset deletes a given asset from the world state.
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
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwnerId string, amount uint64) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	assetCount := asset.Count
	if assetCount < amount {
		return fmt.Errorf("Stock does not have enough count!")
	}
	asset.Count -= amount
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}
	transferredAsset := Asset{asset.Name + "-" + newOwnerId, id, amount, newOwnerId}
	transferredAssetJSON, err := json.Marshal(transferredAsset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(newOwnerId, transferredAssetJSON)
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
