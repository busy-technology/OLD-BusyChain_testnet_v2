package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const balancePrefix = "account~tokenId~sender"
const approvalPrefix = "account~operator"
const tokenAddressPrefix = "0x"

const minterMSPID = "BusyMSP"
const NFT_EVENT = "NFT"

// BusyTokens provides functions for transferring tokens between accounts
type BusyTokens struct {
	contractapi.Contract
}

type TransferSingle struct {
	Operator     string `json:"operator"`
	From         string `json:"from"`
	To           string `json:"to"`
	Symbol       string `json:"symbol"`
	Value        uint64 `json:"value"`
	TokenAddress string `json:"tokenAddress"`
}

type TransferBatch struct {
	Operator       string   `json:"operator"`
	From           string   `json:"from"`
	To             string   `json:"to"`
	Symbols        []string `json:"symbols"`
	Values         []uint64 `json:"values"`
	TokenAddresses []string `json:"tokenAddresses"`
}

type ApprovalForAll struct {
	Owner    string `json:"owner"`
	Operator string `json:"operator"`
	Approved bool   `json:"approved"`
}

// TokenMetaData holds the metadata of the tokens Minted.
type TokenMetaData struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Logo        string      `json:"logo"`
	Properties  interface{} `json:"properties"`
}

// BusyTokensInfo holds metadata and owner info
type BusyTokensInfo struct {
	Account      string        `json:"account"`
	CreatedAT    time.Time     `json:"created_at"`
	TokenAddress string        `json:"tokenAddress"`
	MetaData     TokenMetaData `json:"metadata"`
}

// NFTEvent Holds data for NFT event sent out
type NFTEvent struct {
	UserAddresses UserAddress    `json:"userAddress,omitempty"`
	NFTList       []NFTEventInfo `json:"nftEventInfo"`
}

type NFTEventInfo struct {
	Account   string `json:"account"`
	Symbol    string `json:"symbol"`
	TokenType string `json:"tokenType"`
}

// Mint creates amount tokens of token type and assigns them to account.
func (s *BusyTokens) Mint(ctx contractapi.TransactionContextInterface, account string, symbol string, totalSupply uint64, metadata TokenMetaData) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	tokenAddress := generateTokenAddress(symbol)
	// checking if the token already exists
	busyTokensInfoAsBytes, err := ctx.GetStub().GetState(tokenAddress)
	if err != nil {
		response.Message = fmt.Sprintf("Error while getting state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if busyTokensInfoAsBytes != nil {
		response.Message = "Token already exists"
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// check if token already exists
	exists, err := ifTokenExists(ctx, symbol)
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching token details: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if exists {
		response.Message = fmt.Sprintf("Token with the same symbol %s already exists", symbol)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	if metadata.Logo == "" || metadata.Name == "" || metadata.Type == "" {
		response.Message = "Invalid Metadata"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	if !contains([]string{"NFT", "GAME"}, metadata.Type) {
		response.Message = "Only NFT and GAME are supported as type in metadata"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	// check if wallet already exists
	walletAsBytes, err := ctx.GetStub().GetState(account)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if walletAsBytes == nil {
		response.Message = fmt.Sprintf("Account %s does not exist", account)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// Get Common Name of submitting client identity
	commonName, err := getCommonName(ctx)
	if err != nil {
		response.Message = fmt.Sprintf("failed to get Common name: %v", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	operator, err := getDefaultWalletAddress(ctx, commonName)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	balance, _ := getBalanceHelper(ctx, account, BUSY_COIN_SYMBOL)
	mintFeeString, _ := GetTokenIssueFeeForTokenType(ctx, "nft")
	mintFee, _ := new(big.Int).SetString(mintFeeString, 10)
	if balance.Cmp(mintFee) == -1 {
		response.Message = fmt.Sprintf("User %s does not have the enough balance to mint new tokens", account)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = txFeeHelper(ctx, account, BUSY_COIN_SYMBOL, mintFee.String(), "mint")
	if err != nil {
		response.Message = "error while burning mint Transaction Fee"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	// Mint tokens

	err = mintHelper(ctx, operator, account, symbol, totalSupply)
	if err != nil {
		response.Message = fmt.Sprintf("Error while minting the tokens: %v", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	busyTokensInfo := BusyTokensInfo{
		Account:      account,
		TokenAddress: tokenAddress,
		CreatedAT:    time.Now(),
		MetaData:     metadata,
	}
	// putting the tokenMetaData
	busyTokensInfoAsBytes, _ = json.Marshal(busyTokensInfo)
	err = ctx.GetStub().PutState(tokenAddress, busyTokensInfoAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	// sending the balance event
	nftEventData := NFTEvent{
		UserAddresses: UserAddress{
			Address: account,
			Token:   BUSY_COIN_SYMBOL,
		},
		NFTList: []NFTEventInfo{
			{
				Account:   account,
				TokenType: metadata.Type,
				Symbol:    symbol,
			},
		},
	}
	nftEventDataAsBytes, _ := json.Marshal(nftEventData)
	err = ctx.GetStub().SetEvent(NFT_EVENT, nftEventDataAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while Sending the NFT event: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// Send Success response
	transferSingleData := TransferSingle{operator, "0x", account, symbol, totalSupply, tokenAddress}
	response.Data = transferSingleData
	response.Success = true
	response.Message = "Tokens has been successfully minted"
	return response, nil
}

// MintBatch creates amount tokens for each token type and assigns them to account.
func (s *BusyTokens) MintBatch(ctx contractapi.TransactionContextInterface, account string, symbols []string, totalSupplies []uint64, metadatas []TokenMetaData) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	if len(symbols) != len(totalSupplies) || len(symbols) != len(metadatas) {
		return nil, fmt.Errorf("ids ,amounts and must have the same length")
	}

	// checking if the token already exists
	walletAsBytes, err := ctx.GetStub().GetState(account)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if walletAsBytes == nil {
		response.Message = fmt.Sprintf("Wallet %s does not exist", account)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	tokenAddresses := []string{}
	nftList := []NFTEventInfo{}

	for idx, symbol := range symbols {
		exist, err := ifTokenExists(ctx, symbol)
		if err != nil {
			response.Message = fmt.Sprintf("Error while fetching token details: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		if exist {
			response.Message = fmt.Sprintf("Token with symbol %s already exists", symbol)
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}

		tokenAddress := generateTokenAddress(symbol)
		tokenAddresses = append(tokenAddresses, tokenAddress)
		busyTokensInfoAsBytes, err := ctx.GetStub().GetState(tokenAddress)
		if err != nil {
			response.Message = fmt.Sprintf("Error while getting state in blockchain: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		if busyTokensInfoAsBytes != nil {
			response.Message = "Token already exists"
			logger.Info(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		// putting the tokenMetaData into the state

		if metadatas[idx].Logo == "" || metadatas[idx].Name == "" || metadatas[idx].Type == "" {
			response.Message = "Invalid Metadata"
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}

		busyTokensInfo := BusyTokensInfo{
			Account:      account,
			CreatedAT:    time.Now(),
			TokenAddress: tokenAddress,
			MetaData:     metadatas[idx],
		}
		busyTokensInfoAsBytes, _ = json.Marshal(busyTokensInfo)
		err = ctx.GetStub().PutState(tokenAddress, busyTokensInfoAsBytes)
		if err != nil {
			response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		nftList = append(nftList, NFTEventInfo{
			Account:   account,
			TokenType: metadatas[idx].Type,
			Symbol:    symbol,
		})
	}
	// Get Common Name of submitting client
	commonName, err := getCommonName(ctx)
	if err != nil {
		response.Message = fmt.Sprintf("failed to get Common name: %v", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	operator, err := getDefaultWalletAddress(ctx, commonName)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// Group amount by token symbols because we can only send token to a recipient only one time in a block. This prevents key conflicts
	amountToSend := make(map[string]uint64) // token symbol => amount

	for i := 0; i < len(totalSupplies); i++ {
		amountToSend[symbols[i]] += totalSupplies[i]
	}

	// Copy the map keys and sort it. This is necessary because iterating maps in Go is not deterministic
	amountToSendKeys := sortedKeys(amountToSend)

	// Mint tokens
	for _, id := range amountToSendKeys {
		amount := amountToSend[id]
		err = mintHelper(ctx, operator, account, id, amount)
		if err != nil {
			response.Message = fmt.Sprintf("Error while minting the batch %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
	}
	balance, _ := getBalanceHelper(ctx, account, BUSY_COIN_SYMBOL)
	mintFeeString, _ := GetTokenIssueFeeForTokenType(ctx, "game")
	mintFee, _ := new(big.Int).SetString(mintFeeString, 10)
	numberofTokens := new(big.Int).SetInt64(int64(len(symbols)))
	mintFeeBatch := new(big.Int).Mul(mintFee, numberofTokens)
	if balance.Cmp(mintFeeBatch) == -1 {
		response.Message = fmt.Sprintf("User %s does not have the enough balance to mint tokens", account)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = txFeeHelper(ctx, account, BUSY_COIN_SYMBOL, mintFeeBatch.String(), "mintGame")
	if err != nil {
		response.Message = "error while burning mint Transaction Fee"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// sending the balance event
	nftEventData := NFTEvent{
		UserAddresses: UserAddress{
			Address: account,
			Token:   BUSY_COIN_SYMBOL,
		},
		NFTList: nftList,
	}
	nftEventDataAsBytes, _ := json.Marshal(nftEventData)
	err = ctx.GetStub().SetEvent(NFT_EVENT, nftEventDataAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while Sending the NFT event: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	// Emit TransferBatch event
	transferBatchData := TransferBatch{operator, "0x", account, symbols, totalSupplies, tokenAddresses}
	response.Data = transferBatchData
	response.Message = "Request to mint the tokens has been successfully accepted"
	response.Success = true
	return response, nil
}

// BurnBatch destroys amount tokens of for each token type from account.
// This function emits a TransferBatch event.
func (s *BusyTokens) BurnBatch(ctx contractapi.TransactionContextInterface, account string, symbols []string, amounts []uint64) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}
	if account == "0x" {
		return nil, fmt.Errorf("burn to the zero address")
	}

	if len(symbols) != len(amounts) {
		return nil, fmt.Errorf("ids and amounts must have the same length")
	}

	tokenAddresses := []string{}
	nftList := []NFTEventInfo{}
	for _, symbol := range symbols {
		// checking if the token already exists

		tokenAddress := generateTokenAddress(symbol)
		tokenAddresses = append(tokenAddresses, tokenAddress)
		busyTokensInfoAsBytes, err := ctx.GetStub().GetState(tokenAddress)
		if err != nil {
			response.Message = fmt.Sprintf("Error while getting state in blockchain: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		if busyTokensInfoAsBytes == nil {
			response.Message = fmt.Sprintf("Token %s does not exists", symbol)
			logger.Info(response.Message)
			return response, fmt.Errorf(response.Message)
		}

		busyTokensInfo := BusyTokensInfo{}
		err = json.Unmarshal(busyTokensInfoAsBytes, &busyTokensInfo)
		if err != nil {
			response.Message = fmt.Sprintf("Error while Marshalling the data: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}

		nftList = append(nftList, NFTEventInfo{
			Account:   account,
			Symbol:    symbol,
			TokenType: busyTokensInfo.MetaData.Type,
		})

	}

	walletAsBytes, err := ctx.GetStub().GetState(account)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if walletAsBytes == nil {
		response.Message = fmt.Sprintf("Wallet %s does not exist", account)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// Get Common Name of submitting client identity
	commonName, err := getCommonName(ctx)
	if err != nil {
		response.Message = fmt.Sprintf("failed to get Common name: %v", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if commonName != "busy_network" {
		response.Message = "You are not allowed to Burn the tokens"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	err = removeBalance(ctx, account, symbols, amounts)
	if err != nil {
		response.Message = fmt.Sprintf("Error while burning the tokens: %v", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// sending the balance event
	nftEventData := NFTEvent{
		NFTList: nftList,
	}
	nftEventDataAsBytes, _ := json.Marshal(nftEventData)
	err = ctx.GetStub().SetEvent(NFT_EVENT, nftEventDataAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while Sending the NFT event: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	burnBatchData := TransferBatch{account, account, "0x0", symbols, amounts, tokenAddresses}
	response.Data = burnBatchData
	response.Message = "Tokens burn successfully"
	response.Success = true
	return response, nil
}

// TransferFrom transfers tokens from sender account to recipient account
// recipient account must be a valid clientID as returned by the ClientID() function
// This function triggers a TransferSingle event
func (s *BusyTokens) TransferFrom(ctx contractapi.TransactionContextInterface, sender string, recipient string, symbol string, amount uint64) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}
	if sender == recipient {
		return nil, fmt.Errorf("transfer to self")
	}

	if recipient == "0x" {
		return nil, fmt.Errorf("transfer to the zero address")
	}

	// checking if the token already exists
	tokenAddress := generateTokenAddress(symbol)
	busyTokensInfoAsBytes, err := ctx.GetStub().GetState(tokenAddress)
	if err != nil {
		response.Message = fmt.Sprintf("Error while getting state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if busyTokensInfoAsBytes == nil {
		response.Message = "Token does not exists"
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	busyTokensInfo := BusyTokensInfo{}
	err = json.Unmarshal(busyTokensInfoAsBytes, &busyTokensInfo)
	if err != nil {
		response.Message = fmt.Sprintf("Error while Marshalling the data: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// Get Common Name of submitting client identity
	commonName, err := getCommonName(ctx)
	if err != nil {
		response.Message = fmt.Sprintf("failed to get Common name: %v", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	operator, err := getDefaultWalletAddress(ctx, commonName)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	bigAmount := new(big.Int).SetUint64(amount)
	if bigAmount.Cmp(bigZero) == 0 {
		response.Message = "Amount cannot be zero"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// Check whether operator is owner or approved
	if operator != sender {
		approved, err := _isApprovedForAll(ctx, sender, operator)
		if err != nil {
			response.Message = fmt.Sprintf("failed to get the approval status of operator: %v", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		if !approved {
			response.Message = "Caller is neither the owner nor is approved"
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
	}

	balance, _ := getBalanceHelper(ctx, sender, BUSY_COIN_SYMBOL)
	txFee, _ := getCurrentTxFee(ctx)
	tranferFee, _ := new(big.Int).SetString(txFee, 10)
	if balance.Cmp(tranferFee) == -1 {
		response.Message = fmt.Sprintf("User %s does not have the enough balance to tranfer tokens", sender)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = txFeeHelper(ctx, sender, BUSY_COIN_SYMBOL, tranferFee.String(), "transfer")
	if err != nil {
		response.Message = "error while burning mint Transaction Fee"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// Withdraw the funds from the sender address
	err = removeBalance(ctx, sender, []string{symbol}, []uint64{amount})
	if err != nil {
		response.Message = fmt.Sprintf("Error while removing Balance %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// Deposit the fund to the recipient address
	err = addBalance(ctx, sender, recipient, symbol, amount)
	if err != nil {
		response.Message = "error while adding balance to the recipient"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// sending the balance event
	nftEventData := NFTEvent{
		UserAddresses: UserAddress{
			Address: sender,
			Token:   BUSY_COIN_SYMBOL,
		},
		NFTList: []NFTEventInfo{
			{
				Account:   sender,
				Symbol:    symbol,
				TokenType: busyTokensInfo.MetaData.Type,
			},
			{
				Account:   recipient,
				Symbol:    symbol,
				TokenType: busyTokensInfo.MetaData.Type,
			},
		},
	}
	nftEventDataAsBytes, _ := json.Marshal(nftEventData)
	err = ctx.GetStub().SetEvent(NFT_EVENT, nftEventDataAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while Sending the NFT event: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	transferSingleData := TransferSingle{operator, sender, recipient, symbol, amount, tokenAddress}
	response.Data = transferSingleData
	response.Message = "Request to transfer tokens has been successfully accepted"
	response.Success = true
	return response, nil
}

// BatchTransferFrom transfers multiple tokens from sender account to recipient account
// recipient account must be a valid clientID as returned by the ClientID() function
// This function triggers a TransferBatch event
func (s *BusyTokens) BatchTransferFrom(ctx contractapi.TransactionContextInterface, sender string, recipient string, symbols []string, amounts []uint64) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}
	if sender == recipient {
		return nil, fmt.Errorf("transfer to self")
	}

	if len(symbols) != len(amounts) {
		return nil, fmt.Errorf("ids and amounts must have the same length")
	}
	if recipient == "0x" {
		return nil, fmt.Errorf("transfer to the zero address")
	}

	// Get Common Name of submitting client identity
	commonName, err := getCommonName(ctx)
	if err != nil {
		response.Message = fmt.Sprintf("failed to get Common name: %v", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	operator, err := getDefaultWalletAddress(ctx, commonName)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// Check whether operator is owner or approved
	if operator != sender {
		approved, err := _isApprovedForAll(ctx, sender, operator)
		if err != nil {
			response.Message = fmt.Sprintf("failed to get the approval for operator: %v", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		if !approved {
			response.Message = "caller is not owner nor is approved"
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
	}

	tokenAddresses := []string{}
	nftList := []NFTEventInfo{}
	for idx := range symbols {
		// checking if the token already exists
		tokenAddress := generateTokenAddress(symbols[idx])
		tokenAddresses = append(tokenAddresses, tokenAddress)
		busyTokensInfoAsBytes, err := ctx.GetStub().GetState(tokenAddress)
		if err != nil {
			response.Message = fmt.Sprintf("Error while getting state in blockchain: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		if busyTokensInfoAsBytes == nil {
			response.Message = "Token does not exists"
			logger.Info(response.Message)
			return response, fmt.Errorf(response.Message)
		}

		busyTokensInfo := BusyTokensInfo{}
		err = json.Unmarshal(busyTokensInfoAsBytes, &busyTokensInfo)
		if err != nil {
			response.Message = fmt.Sprintf("Error while Marshalling the data: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}

		nftList = append(nftList, NFTEventInfo{
			Account:   sender,
			TokenType: busyTokensInfo.MetaData.Type,
			Symbol:    symbols[idx],
		})
		nftList = append(nftList, NFTEventInfo{
			Account:   recipient,
			TokenType: busyTokensInfo.MetaData.Type,
			Symbol:    symbols[idx],
		})
		bigAmount := new(big.Int).SetUint64(amounts[idx])
		if bigAmount.Cmp(bigZero) == 0 {
			response.Message = "Amount cannot be zero"
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}

	}

	balance, _ := getBalanceHelper(ctx, sender, BUSY_COIN_SYMBOL)
	txFee, _ := getCurrentTxFee(ctx)
	tranferFee, _ := new(big.Int).SetString(txFee, 10)
	numberofTokens := new(big.Int).SetInt64(int64(len(symbols)))
	transferFeeBatch := new(big.Int).Mul(tranferFee, numberofTokens)
	if balance.Cmp(transferFeeBatch) == -1 {
		response.Message = fmt.Sprintf("User %s does not have the enough balance to tranfer tokens", sender)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = txFeeHelper(ctx, sender, BUSY_COIN_SYMBOL, transferFeeBatch.String(), "transferBatch")
	if err != nil {
		response.Message = "error while burning mint Transaction Fee"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// Withdraw the funds from the sender address
	err = removeBalance(ctx, sender, symbols, amounts)
	if err != nil {
		response.Message = fmt.Sprintf("Error while removing the balance %s", err.Error())
		logger.Error(response.Message)
		return nil, err
	}

	// Group amount by token symbols because we can only send token to a recipient only one time in a block. This prevents key conflicts
	amountToSend := make(map[string]uint64) // token symbol => amount

	for i := 0; i < len(amounts); i++ {
		amountToSend[symbols[i]] += amounts[i]
	}

	// Copy the map keys and sort it. This is necessary because iterating maps in Go is not deterministic
	amountToSendKeys := sortedKeys(amountToSend)

	// Deposit the funds to the recipient address
	for _, id := range amountToSendKeys {
		amount := amountToSend[id]
		err = addBalance(ctx, sender, recipient, id, amount)
		if err != nil {
			response.Message = fmt.Sprintf("Error while adding the balance to the recipient %s", err.Error())
			logger.Error(response.Message)
			return response, err
		}
	}

	// sending the balance event
	nftEventData := NFTEvent{
		UserAddresses: UserAddress{
			Address: sender,
			Token:   BUSY_COIN_SYMBOL,
		},
		NFTList: nftList,
	}
	nftEventDataAsBytes, _ := json.Marshal(nftEventData)
	err = ctx.GetStub().SetEvent(NFT_EVENT, nftEventDataAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while Sending the NFT event: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	transferBatchData := TransferBatch{operator, sender, recipient, symbols, amounts, tokenAddresses}
	response.Data = transferBatchData
	response.Message = "Request to transfer tokens has been successfully accepted"
	response.Success = true
	return response, nil
}

// IsApprovedForAll returns true if operator is approved to transfer account's tokens.
func (s *BusyTokens) IsApprovedForAll(ctx contractapi.TransactionContextInterface, account string, operator string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	// check if operator does not exists
	walletAsBytes, err := ctx.GetStub().GetState(operator)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if walletAsBytes == nil {
		response.Message = fmt.Sprintf("Operator %s does not exist", operator)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	isApproved, err := _isApprovedForAll(ctx, account, operator)
	if err != nil {
		return response, err
	}
	response.Data = isApproved
	response.Message = "The operator's approval status has been successfully fetched"
	response.Success = true
	return response, nil
}

// _isApprovedForAll returns true if operator is approved to transfer account's tokens.
func _isApprovedForAll(ctx contractapi.TransactionContextInterface, account string, operator string) (bool, error) {
	approvalKey, err := ctx.GetStub().CreateCompositeKey(approvalPrefix, []string{account, operator})
	if err != nil {
		return false, fmt.Errorf("failed to create the composite key for prefix %s: %v", approvalPrefix, err)
	}

	approvalBytes, err := ctx.GetStub().GetState(approvalKey)
	if err != nil {
		return false, fmt.Errorf("failed to read approval of operator %s for account %s from world state: %v", operator, account, err)
	}

	if approvalBytes == nil {
		return false, nil
	}

	var approved bool
	err = json.Unmarshal(approvalBytes, &approved)
	if err != nil {
		return false, fmt.Errorf("failed to decode approval JSON of operator %s for account %s: %v", operator, account, err)
	}

	return approved, nil
}

// SetApprovalForAll returns true if operator is approved to transfer account's tokens.
func (s *BusyTokens) SetApprovalForAll(ctx contractapi.TransactionContextInterface, operator string, approved bool) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}
	// Get Common Name of submitting client identity
	commonName, err := getCommonName(ctx)
	if err != nil {
		response.Message = fmt.Sprintf("failed to get Common name: %v", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	account, err := getDefaultWalletAddress(ctx, commonName)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	if account == operator {
		return nil, fmt.Errorf("setting approval status for self")
	}

	// check if operator does not exists
	walletAsBytes, err := ctx.GetStub().GetState(operator)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if walletAsBytes == nil {
		response.Message = fmt.Sprintf("Operator %s does not exist", operator)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	balance, _ := getBalanceHelper(ctx, account, BUSY_COIN_SYMBOL)
	txFee, _ := getCurrentTxFee(ctx)
	bigTxFee, _ := new(big.Int).SetString(txFee, 10)
	if balance.Cmp(bigTxFee) == -1 {
		response.Message = fmt.Sprintf("User %s does not have the enough balance to Set Approval for NFT", account)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = txFeeHelper(ctx, account, BUSY_COIN_SYMBOL, bigTxFee.String(), "busynftTransfer")
	if err != nil {
		response.Message = "Error while burning Transaction Fee"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	approvalKey, err := ctx.GetStub().CreateCompositeKey(approvalPrefix, []string{account, operator})
	if err != nil {
		response.Message = fmt.Sprintf("failed to create the composite key for prefix %s: %v", approvalPrefix, err)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	approvalJSON, err := json.Marshal(approved)
	if err != nil {
		response.Message = fmt.Sprintf("failed to encode approval JSON of operator %s for account %s: %v", operator, account, err)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	err = ctx.GetStub().PutState(approvalKey, approvalJSON)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	balanceData := []UserAddress{
		{
			Address: account,
			Token:   BUSY_COIN_SYMBOL,
		},
	}
	balanceAsBytes, _ := json.Marshal(balanceData)
	err = ctx.GetStub().SetEvent(BALANCE_EVENT, balanceAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while Sending the Balance event: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	approvalForAllData := ApprovalForAll{account, operator, approved}
	response.Data = approvalForAllData
	response.Message = "Request to set approval has been successfully accepted"
	response.Success = true
	return response, nil
}

// BalanceOf returns the balance of the given account
func (s *BusyTokens) BalanceOf(ctx contractapi.TransactionContextInterface, account string, symbol string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	// checking if the token already exists
	tokenAddress := generateTokenAddress(symbol)
	metaDateAsBytes, err := ctx.GetStub().GetState(tokenAddress)
	if err != nil {
		response.Message = fmt.Sprintf("Error while getting state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if metaDateAsBytes == nil {
		response.Message = fmt.Sprintf("Token %s does not exist", symbol)
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	balance, err := balanceOfHelper(ctx, account, symbol)
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching the balance %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	response.Data = balance
	response.Message = "Balance of the token has been successfully fetched"
	response.Success = true
	return response, nil
}

// BalanceOfBatch returns the balance of multiple account/token pairs
func (s *BusyTokens) BalanceOfBatch(ctx contractapi.TransactionContextInterface, accounts []string, symbols []string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}
	if len(accounts) != len(symbols) {
		return nil, fmt.Errorf("accounts and ids must have the same length")
	}

	balances := make([]uint64, len(accounts))

	for i := 0; i < len(accounts); i++ {
		var err error
		// checking if the token already exists
		tokenAddress := generateTokenAddress(symbols[i])
		metaDateAsBytes, err := ctx.GetStub().GetState(tokenAddress)
		if err != nil {
			response.Message = fmt.Sprintf("Error while getting state in blockchain: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		if metaDateAsBytes == nil {
			response.Message = fmt.Sprintf("Token %s does not exist", symbols[i])
			logger.Info(response.Message)
			return response, fmt.Errorf(response.Message)
		}

		// check if wallet already exists
		walletAsBytes, err := ctx.GetStub().GetState(accounts[i])
		if err != nil {
			response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		if walletAsBytes == nil {
			response.Message = fmt.Sprintf("Account %s does not exist", accounts[i])
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		balances[i], err = balanceOfHelper(ctx, accounts[i], symbols[i])
		if err != nil {
			response.Message = fmt.Sprintf("Failed to fetch the balance %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
	}

	response.Data = balances
	response.Message = "Balance of the tokens has been successfully fetched"
	response.Success = true
	return response, nil
}

// GetTokenInfo returns the metadata, owner and tokenAddress of the Requested token
func (s *BusyTokens) GetTokenInfo(ctx contractapi.TransactionContextInterface, symbol string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	// putting the tokenMetaData
	tokenAddress := generateTokenAddress(symbol)
	busyTokensInfoAsBytes, err := ctx.GetStub().GetState(tokenAddress)
	if err != nil {
		response.Message = fmt.Sprintf("Error while getting state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if busyTokensInfoAsBytes == nil {
		response.Message = "Token does not exist"
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	busyTokensInfo := BusyTokensInfo{}
	err = json.Unmarshal(busyTokensInfoAsBytes, &busyTokensInfo)
	if err != nil {
		response.Message = fmt.Sprintf("Error while Marshalling the data: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	response.Data = busyTokensInfo
	response.Message = "The tokens Info has been successfully fetched"
	response.Success = true
	return response, nil
}

// UpdateTokenMetaData to update TokenMetaData
func (s *BusyTokens) UpdateTokenMetaData(ctx contractapi.TransactionContextInterface, symbol string, metadata TokenMetaData) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	// putting the tokenMetaData
	tokenAddress := generateTokenAddress(symbol)
	busyTokensInfoAsBytes, err := ctx.GetStub().GetState(tokenAddress)
	if err != nil {
		response.Message = fmt.Sprintf("Error while getting state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if busyTokensInfoAsBytes == nil {
		response.Message = "Token does not exist"
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	busyTokensInfo := BusyTokensInfo{}
	err = json.Unmarshal(busyTokensInfoAsBytes, &busyTokensInfo)
	if err != nil {
		response.Message = fmt.Sprintf("Error while Marshalling the data: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// Get Common Name of submitting client identity
	commonName, err := getCommonName(ctx)
	if err != nil {
		response.Message = fmt.Sprintf("failed to get Common name: %v", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	defaultWalletAddress, err := getDefaultWalletAddress(ctx, commonName)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	if defaultWalletAddress != busyTokensInfo.Account {
		response.Message = fmt.Sprintf("The account %s is not the owner of %s", defaultWalletAddress, symbol)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	balance, _ := getBalanceHelper(ctx, defaultWalletAddress, BUSY_COIN_SYMBOL)
	txFee, _ := getCurrentTxFee(ctx)
	bigTxFee, _ := new(big.Int).SetString(txFee, 10)
	if balance.Cmp(bigTxFee) == -1 {
		response.Message = fmt.Sprintf("User %s does not have the enough balance to Update Metadata of NFT", defaultWalletAddress)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = txFeeHelper(ctx, defaultWalletAddress, BUSY_COIN_SYMBOL, bigTxFee.String(), "busynftTransfer")
	if err != nil {
		response.Message = "Error while burning Transaction Fee"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	if busyTokensInfo.MetaData.Type != metadata.Type {
		response.Message = "Token Type cannot be updated"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	busyTokensInfo.MetaData = metadata

	// unmarshall and putting in state
	busyTokensInfoAsBytes, _ = json.Marshal(busyTokensInfo)
	err = ctx.GetStub().PutState(tokenAddress, busyTokensInfoAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	balanceData := []UserAddress{
		{
			Address: defaultWalletAddress,
			Token:   BUSY_COIN_SYMBOL,
		},
	}
	balanceAsBytes, _ := json.Marshal(balanceData)
	err = ctx.GetStub().SetEvent(BALANCE_EVENT, balanceAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while Sending the Balance event: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	response.Data = metadata
	response.Message = "The token's metadata has been successfully updated"
	response.Success = true
	return response, nil
}
func mintHelper(ctx contractapi.TransactionContextInterface, operator string, account string, symbol string, amount uint64) error {
	if account == "0x" {
		return fmt.Errorf("mint to the zero address")
	}

	if amount <= 0 {
		return fmt.Errorf("mint amount must be a positive integer")
	}

	err := addBalance(ctx, operator, account, symbol, amount)
	if err != nil {
		return err
	}

	return nil
}

func addBalance(ctx contractapi.TransactionContextInterface, sender string, recipient string, symbol string, amount uint64) error {

	balanceKey, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{recipient, symbol, sender})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", balancePrefix, err)
	}

	balanceBytes, err := ctx.GetStub().GetState(balanceKey)
	if err != nil {
		return fmt.Errorf("failed to read account %s from world state: %v", recipient, err)
	}

	var balance uint64 = 0
	if balanceBytes != nil {
		balance, _ = strconv.ParseUint(string(balanceBytes), 10, 64)
	}

	balance += amount

	err = ctx.GetStub().PutState(balanceKey, []byte(strconv.FormatUint(uint64(balance), 10)))
	if err != nil {
		return err
	}

	return nil
}

func setBalance(ctx contractapi.TransactionContextInterface, sender string, recipient string, symbol string, amount uint64) error {

	balanceKey, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{recipient, symbol, sender})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", balancePrefix, err)
	}

	err = ctx.GetStub().PutState(balanceKey, []byte(strconv.FormatUint(uint64(amount), 10)))
	if err != nil {
		return err
	}

	return nil
}

func removeBalance(ctx contractapi.TransactionContextInterface, sender string, symbols []string, amounts []uint64) error {
	// Calculate the total amount of each token to withdraw
	necessaryFunds := make(map[string]uint64) // token symbol -> necessary amount

	for i := 0; i < len(amounts); i++ {
		necessaryFunds[symbols[i]] += amounts[i]
	}

	// Copy the map keys and sort it. This is necessary because iterating maps in Go is not deterministic
	necessaryFundsKeys := sortedKeys(necessaryFunds)

	// Check whether the sender has the necessary funds and withdraw them from the account
	for _, tokenId := range necessaryFundsKeys {
		neededAmount := necessaryFunds[tokenId]

		var partialBalance uint64
		var selfRecipientKeyNeedsToBeRemoved bool
		var selfRecipientKey string

		balanceIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(balancePrefix, []string{sender, tokenId})
		if err != nil {
			return fmt.Errorf("failed to get state for prefix %v: %v", balancePrefix, err)
		}
		defer balanceIterator.Close()

		// Iterate over keys that store balances and add them to partialBalance until
		// either the necessary amount is reached or the keys ended
		for balanceIterator.HasNext() && partialBalance < neededAmount {
			queryResponse, err := balanceIterator.Next()
			if err != nil {
				return fmt.Errorf("failed to get the next state for prefix %v: %v", balancePrefix, err)
			}

			partBalAmount, _ := strconv.ParseUint(string(queryResponse.Value), 10, 64)
			partialBalance += partBalAmount

			_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
			if err != nil {
				return err
			}

			if compositeKeyParts[2] == sender {
				selfRecipientKeyNeedsToBeRemoved = true
				selfRecipientKey = queryResponse.Key
			} else {
				err = ctx.GetStub().DelState(queryResponse.Key)
				if err != nil {
					return fmt.Errorf("failed to delete the state of %v: %v", queryResponse.Key, err)
				}
			}
		}

		if partialBalance < neededAmount {
			return fmt.Errorf("sender has insufficient funds for token %v, needed funds: %v, available fund: %v", tokenId, neededAmount, partialBalance)
		} else if partialBalance > neededAmount {
			// Send the remainder back to the sender
			remainder := partialBalance - neededAmount
			if selfRecipientKeyNeedsToBeRemoved {
				// Set balance for the key that has the same address for sender and recipient
				err = setBalance(ctx, sender, sender, tokenId, remainder)
				if err != nil {
					return err
				}
			} else {
				err = addBalance(ctx, sender, sender, tokenId, remainder)
				if err != nil {
					return err
				}
			}

		} else {
			// Delete self recipient key
			err = ctx.GetStub().DelState(selfRecipientKey)
			if err != nil {
				return fmt.Errorf("failed to delete the state of %v: %v", selfRecipientKey, err)
			}
		}
	}

	return nil
}

// balanceOfHelper returns the balance of the given account
func balanceOfHelper(ctx contractapi.TransactionContextInterface, account string, symbol string) (uint64, error) {

	if account == "0x" {
		return 0, fmt.Errorf("balance query for the zero address")
	}

	var balance uint64

	balanceIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(balancePrefix, []string{account, symbol})
	if err != nil {
		return 0, fmt.Errorf("failed to get state for prefix %v: %v", balancePrefix, err)
	}
	defer balanceIterator.Close()

	for balanceIterator.HasNext() {
		queryResponse, err := balanceIterator.Next()
		if err != nil {
			return 0, fmt.Errorf("failed to get the next state for prefix %v: %v", balancePrefix, err)
		}

		balAmount, _ := strconv.ParseUint(string(queryResponse.Value), 10, 64)
		balance += balAmount
	}

	return balance, nil
}

// Returns the sorted slice ([]uint64) copied from the keys of map[uint64]uint64
func sortedKeys(m map[string]uint64) []string {
	// Copy map keys to slice
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	// Sort the slice
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// txFeeHelper burns fee from the user and reduce total supply accordingly
func txFeeHelper(ctx contractapi.TransactionContextInterface, address string, token string, txFee string, txType string) error {
	minusOne, _ := new(big.Int).SetString("-1", 10)
	bigTxFee, _ := new(big.Int).SetString(txFee, 10)
	err := addTotalSupplyUTXO(ctx, token, new(big.Int).Set(bigTxFee).Mul(minusOne, bigTxFee))
	if err != nil {
		return err
	}

	// err = addUTXO(ctx, address, bigTxFee, token)
	// if err != nil {
	// 	return err
	// }
	utxo := UTXO{
		DocType: "utxo",
		Address: address,
		Amount:  bigTxFee.Mul(bigTxFee, minusOne).String(),
		Token:   BUSY_COIN_SYMBOL,
	}
	utxoAsBytes, _ := json.Marshal(utxo)
	err = ctx.GetStub().PutState(fmt.Sprintf("burnTxFee~%s~%s~%s~%s", ctx.GetStub().GetTxID(), txType, address, BUSY_COIN_SYMBOL), utxoAsBytes)
	if err != nil {
		return err
	}
	return nil
}

// check if string is in slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func generateTokenAddress(symbol string) string {
	return "B-" + tokenAddressPrefix + base64Encode(fmt.Sprintf("token-meta-%s", symbol))
}

func base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// func base64Decode(str string) (string, bool) {
// 	data, err := base64.StdEncoding.DecodeString(str)
// 	if err != nil {
// 		return "", true
// 	}
// 	return string(data), false
// }
