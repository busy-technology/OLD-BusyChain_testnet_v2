package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric/common/flogging"
)

// Busy chaincode
type Busy struct {
	contractapi.Contract
}

var logger = flogging.MustGetLogger(BUSY_COIN_SYMBOL)

const TRANSFER_FEE string = "1000000000000000"
const PHASE1_STAKING_AMOUNT = "1000000000000000000000"
const BALANCE_EVENT = "BALANCE"

// Init Initialise chaincocode while deployment
func (bt *Busy) Init(ctx contractapi.TransactionContextInterface) Response {
	response := Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	mspid, _ := ctx.GetClientIdentity().GetMSPID()
	if mspid != "BusyMSP" {
		response.Message = "You are not allowed to issue busy coin"
		logger.Error(response.Message)
		return response
	}
	commonName, _ := getCommonName(ctx)
	if commonName != "busy_network" {
		response.Message = "You are not allowed to issue busy coin"
		logger.Error(response.Message)
		return response
	}
	// setting Message Config
	config := MessageConfig{
		BigBusyCoins:    "1000000000000000000",
		BusyCoin:        1,
		MessageInterval: 5 * time.Second,
	}
	configAsBytes, _ := json.Marshal(config)
	err := ctx.GetStub().PutState("MessageConfig", configAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response
	}

	tokenIssueFeeConfig := TokenIssueFee{
		Busy20: "1500000000000000000000",
		NFT:    "1500000000000000000000",
		Game:   "1500000000000000000000",
	}
	configAsBytes, _ = json.Marshal(tokenIssueFeeConfig)
	err = ctx.GetStub().PutState("TokenIssueFees", configAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating  token issue fees state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response
	}

	// setting Voting Config
	votingConfig := VotingConfig{
		MinimumCoins:    "3400000000000000000000",
		PoolFee:         "1700000000000000000000",
		VotingPeriod:    25 * 60 * time.Minute,
		VotingStartTime: 5 * 60 * time.Minute,
	}
	votingConfigAsBytes, _ := json.Marshal(votingConfig)
	err = ctx.GetStub().PutState("VotingConfig", votingConfigAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response
	}
	now, _ := ctx.GetStub().GetTxTimestamp()

	supply, _ := new(big.Int).SetString("255000000000000000000000000", 10)
	token := Token{
		DocType:     "token",
		ID:          0,
		TokenName:   "Busy",
		TokenSymbol: BUSY_COIN_SYMBOL,
		Admin:       commonName,
		TotalSupply: supply.String(),
		Decimals:    18,
	}
	tokenAsBytes, _ := json.Marshal(token)
	err = ctx.GetStub().PutState(generateTokenStateAddress(BUSY_COIN_SYMBOL), tokenAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating coin on blockchain : %s", err.Error())
		logger.Error(response.Message)
		return response
	}

	err = addTotalSupplyUTXO(ctx, BUSY_COIN_SYMBOL, supply)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating coin on blockchain : %s", err.Error())
		logger.Error(response.Message)
		return response
	}

	wallet := Wallet{
		DocType:   "wallet",
		UserID:    commonName,
		Address:   response.TxID,
		Balance:   supply.String(),
		CreatedAt: uint64(now.Seconds),
	}
	walletAsBytes, _ := json.Marshal(wallet)
	err = ctx.GetStub().PutState(response.TxID, walletAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response
	}
	_ = ctx.GetStub().PutState("latestTokenId", []byte(strconv.Itoa(0)))
	user := User{
		DocType:       "user",
		UserID:        commonName,
		DefaultWallet: wallet.Address,
	}
	userAsBytes, _ := json.Marshal(user)
	err = ctx.GetStub().PutState(commonName, userAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response
	}

	utxo := UTXO{
		DocType: "utxo",
		Address: wallet.Address,
		Amount:  supply.String(),
		Token:   BUSY_COIN_SYMBOL,
	}
	utxoAsBytes, _ := json.Marshal(utxo)
	err = ctx.GetStub().PutState(fmt.Sprintf("%s~%s~%s", response.TxID, wallet.Address, token.TokenSymbol), utxoAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response
	}

	err = ctx.GetStub().PutState("transferFees", []byte(TRANSFER_FEE))
	if err != nil {
		response.Message = fmt.Sprintf("Error while configuring transfer fee: %s", err.Error())
		logger.Error(response.Message)
		return response
	}

	currentStakingLimit, _ := new(big.Int).SetString(PHASE1_STAKING_AMOUNT, 10)
	phaseConfig := PhaseConfig{
		CurrentPhase:          1,
		TotalStakingAddr:      bigZero.String(),
		NextStakingAddrTarget: "100",
		CurrentStakingLimit:   currentStakingLimit.String(),
	}
	phaseConfigAsBytes, _ := json.Marshal(phaseConfig)
	err = ctx.GetStub().PutState("phaseConfig", phaseConfigAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while initialising phase config: %s", err.Error())
		logger.Error(response.Message)
		return response
	}

	phaseUpdateTimeline := map[uint64]PhaseUpdateInfo{
		1: PhaseUpdateInfo{
			UpdatedAt:    uint64(now.Seconds),
			StakingLimit: phaseConfig.CurrentStakingLimit,
		},
	}
	// phaseUpdateTimelineAsBytes, err := ctx.GetStub().GetState(PHASE_UPDATE_TIMELINE)
	// _ = json.Unmarshal(phaseUpdateTimelineAsBytes, &phaseUpdateTimeline)
	// phaseUpdateTimeline[phaseConfig.CurrentPhase] = uint64(now.Seconds)
	phaseUpdateTimelineAsBytes, _ := json.Marshal(phaseUpdateTimeline)
	err = ctx.GetStub().PutState(PHASE_UPDATE_TIMELINE, phaseUpdateTimelineAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while initialising phase timeline: %s", err.Error())
		logger.Error(response.Message)
		return response
	}

	response.Message = fmt.Sprintf("Successfully issued %s", BUSY_COIN_SYMBOL)
	response.Success = true
	response.Data = token
	logger.Info(response.Message)
	return response
}

// CreateUser creates new user on busy blockchain
func (bt *Busy) CreateUser(ctx contractapi.TransactionContextInterface) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	commonName, _ := getCommonName(ctx)
	userAsBytes, err := ctx.GetStub().GetState(commonName)
	if userAsBytes != nil {
		//response.Message = fmt.Sprintf("User nickname is already takenUser with common name %s already exists", commonName)
		response.Message = "User nickname is already taken"
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching user from blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	now, _ := ctx.GetStub().GetTxTimestamp()

	wallet := Wallet{
		DocType:   "wallet",
		UserID:    commonName,
		Address:   "B-" + response.TxID,
		CreatedAt: uint64(now.Seconds),
		// Balance: 0.00,
	}
	walletAsBytes, _ := json.Marshal(wallet)
	err = ctx.GetStub().PutState("B-"+response.TxID, walletAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// userID, _ := ctx.GetClientIdentity().GetID()
	user := User{
		DocType:       "user",
		UserID:        commonName,
		DefaultWallet: wallet.Address,
		MessageCoins: map[string]int{
			"totalCoins": 0,
		},
	}
	userAsBytes, _ = json.Marshal(user)
	err = ctx.GetStub().PutState(commonName, userAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	//response.Message = fmt.Sprintf("User %s has been successfully registered", commonName)
	response.Message = "User has been successfully registered"
	response.Success = true
	response.Data = wallet.Address
	logger.Info(response.Message)
	return response, nil
}

// CreateStakingAddress create new staking address for user
func (bt *Busy) CreateStakingAddress(ctx contractapi.TransactionContextInterface) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	// utxo := UTXO{
	// 	Address: response.TxID,
	// 	Amount:  0.0,
	// 	Token:   "Busy",
	// }
	// utxoAsBytes, _ := json.Marshal(utxo)
	// err := ctx.GetStub().PutState(response.TxID, utxoAsBytes)
	// if err != nil {
	// 	response.Message = fmt.Sprintf("Error while creating new staking address: %s" + err.Error())
	// 	logger.Error(response.Message)
	// 	return response
	// }
	currentPhaseConfig, err := getPhaseConfig(ctx)
	fmt.Println(currentPhaseConfig)
	if err != nil {
		response.Message = fmt.Sprintf("Error while getting phase config: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	now, _ := ctx.GetStub().GetTxTimestamp()

	fmt.Println(currentPhaseConfig.CurrentStakingLimit)
	stakingAmount, _ := new(big.Int).SetString(currentPhaseConfig.CurrentStakingLimit, 10)
	// bigZero, _ := new(big.Int).SetString("0", 10)
	commonName, _ := getCommonName(ctx)
	defaultWalletAddress, _ := getDefaultWalletAddress(ctx, commonName)

	balance, err := getBalanceHelper(ctx, defaultWalletAddress, BUSY_COIN_SYMBOL)
	if balance.Cmp(stakingAmount) == -1 {
		//response.Message = fmt.Sprintf("User %s does not enough coins %s to create a staking address", commonName, currentPhaseConfig.CurrentStakingLimit)
		response.Message = "You do not have enough coins to create a staking address"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	stakingAddress := Wallet{
		DocType:   "stakingAddr",
		UserID:    commonName,
		Address:   "staking-" + response.TxID,
		Balance:   stakingAmount.String(),
		CreatedAt: uint64(now.Seconds),
	}
	txFee, err := getCurrentTxFee(ctx)
	bigTxFee, _ := new(big.Int).SetString(txFee, 10)
	if err != nil {
		response.Message = fmt.Sprintf("Error while getting tx fee: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = transferHelper(ctx, defaultWalletAddress, stakingAddress.Address, stakingAmount, BUSY_COIN_SYMBOL, new(big.Int).Set(bigTxFee))
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while transfering coins to the staking address: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = addTotalSupplyUTXO(ctx, BUSY_COIN_SYMBOL, new(big.Int).Set(bigTxFee).Mul(minusOne, bigTxFee))
	if err != nil {
		response.Message = fmt.Sprintf("Error while burning transfer fee: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	stakingAddrAsBytes, _ := json.Marshal(stakingAddress)
	err = ctx.GetStub().PutState("staking-"+response.TxID, stakingAddrAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	stakingInfo := StakingInfo{
		DocType:              "stakingInfo",
		StakingAddress:       stakingAddress.Address,
		Amount:               stakingAddress.Balance,
		TimeStamp:            uint64(now.Seconds),
		Phase:                currentPhaseConfig.CurrentPhase,
		TotalReward:          bigZero.String(),
		Claimed:              bigZero.String(),
		DefaultWalletAddress: defaultWalletAddress,
	}
	stakingInfoAsBytes, _ := json.Marshal(stakingInfo)
	err = ctx.GetStub().PutState(fmt.Sprintf("info~%s", stakingAddress.Address), stakingInfoAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating staking information in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	_, err = updatePhase(ctx)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating phase: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	// Sending Balance Event
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
	response.Success = true
	response.Message = "Staking address has been successfully created"
	response.Data = stakingInfo
	logger.Info(response.Message)
	return response, nil
}

// GetBalance of specified wallet address
func (bt *Busy) GetBalance(ctx contractapi.TransactionContextInterface, address string, token string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	if token == "" {
		token = BUSY_COIN_SYMBOL
	}
	exists, err := ifTokenExists(ctx, token)
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching token details: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if !exists {
		response.Message = fmt.Sprintf("Symbol %s does not exist", token)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	balance, err := getBalanceHelper(ctx, address, token)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching balance: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	response.Message = "Balance has been successfully fetched"
	response.Success = true
	response.Data = balance.String()
	logger.Info(response.Message)
	return response, nil
}

// GetUser all the wallet and staking address of user with it's balance
func (bt *Busy) GetUser(ctx contractapi.TransactionContextInterface, userID string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	userAsBytes, err := ctx.GetStub().GetState(userID)
	if userAsBytes == nil {
		response.Message = fmt.Sprintf("User %s does not exist", userID)
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching user from blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	userDetails := User{}
	if err := json.Unmarshal(userAsBytes, &userDetails); err != nil {
		response.Message = fmt.Sprintf("Error while retrieving the sender details %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	fmt.Println(userDetails)

	var queryString string = fmt.Sprintf(`{
		"selector": {
			"userId": "%s",
			"docType": "stakingAddr"
		 } 
	}`, userID)
	resultIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching user wallets: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	defer resultIterator.Close()

	var wallet Wallet
	responseData := map[string]interface{}{}
	for resultIterator.HasNext() {
		data, _ := resultIterator.Next()
		json.Unmarshal(data.Value, &wallet)
		balance, _ := getBalanceHelper(ctx, wallet.Address, BUSY_COIN_SYMBOL)
		walletDetails := make(map[string]interface{}, 3)
		walletDetails["balance"] = balance.String()
		walletDetails["token"] = BUSY_COIN_SYMBOL
		walletDetails["createdAt"] = wallet.CreatedAt
		responseData[wallet.Address] = walletDetails
	}

	defaultWalletDetails := make(map[string]interface{}, 3)

	walletAsBytes, err := ctx.GetStub().GetState(userDetails.DefaultWallet)
	if err := json.Unmarshal(walletAsBytes, &wallet); err != nil {
		response.Message = fmt.Sprintf("Error while retrieving the sender details %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	balance, _ := getBalanceHelper(ctx, userDetails.DefaultWallet, BUSY_COIN_SYMBOL)
	defaultWalletDetails["balance"] = balance.String()
	defaultWalletDetails["token"] = BUSY_COIN_SYMBOL
	defaultWalletDetails["createdAt"] = wallet.CreatedAt

	responseData[userDetails.DefaultWallet] = defaultWalletDetails

	responseData["messageCoins"] = userDetails.MessageCoins
	response.Message = "Balance has been successfully fetched"
	response.Success = true
	response.Data = responseData
	logger.Info(response.Message)
	return response, nil
}

func (bt *Busy) GetTokenIssueFee(ctx contractapi.TransactionContextInterface) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	feesAsBytes, err := ctx.GetStub().GetState("TokenIssueFees")
	if feesAsBytes == nil {
		response.Message = fmt.Sprintf(" TokenIssueFees does not exist")
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching TokenIssueFees from blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	tokenIssueFees := TokenIssueFee{}
	if err := json.Unmarshal(feesAsBytes, &tokenIssueFees); err != nil {
		response.Message = fmt.Sprintf("Error while retrieving TokenIssueFees %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	response.Data = tokenIssueFees
	response.Message = "Current token issue fee fee has been successfully fetched"
	response.Success = true
	return response, nil
}

func GetTokenIssueFeeForTokenType(ctx contractapi.TransactionContextInterface, tokenType string) (string, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	feesAsBytes, err := ctx.GetStub().GetState("TokenIssueFees")
	if feesAsBytes == nil {
		response.Message = fmt.Sprintf(" TokenIssueFees does not exist")
		logger.Info(response.Message)
		return "", fmt.Errorf(response.Message)
	}
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching user from blockchain: %s", err.Error())
		logger.Error(response.Message)
		return "", fmt.Errorf(response.Message)
	}

	tokenIssueFees := TokenIssueFee{}
	if err := json.Unmarshal(feesAsBytes, &tokenIssueFees); err != nil {
		response.Message = fmt.Sprintf("Error while retrieving the sender details %s", err.Error())
		logger.Error(response.Message)
		return "", fmt.Errorf(response.Message)
	}

	if strings.ToLower(tokenType) == "busy20" {
		return tokenIssueFees.Busy20, nil
	} else if (strings.ToLower(tokenType) == "nft") || (strings.ToLower(tokenType) == "game") {
		return tokenIssueFees.NFT, nil
	} else {
		return "", fmt.Errorf("Invalid token type , please send token type from [Busy20, NFT,Game]")
	}
}

func (bt *Busy) SetTokenIssueFee(ctx contractapi.TransactionContextInterface, tokenType string, newFee string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	feesAsBytes, err := ctx.GetStub().GetState("TokenIssueFees")
	if feesAsBytes == nil {
		response.Message = fmt.Sprintf(" TokenIssueFees does not exist")
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching TokenIssueFees from blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	tokenIssueFees := TokenIssueFee{}
	if err := json.Unmarshal(feesAsBytes, &tokenIssueFees); err != nil {
		response.Message = fmt.Sprintf("Error while retrieving the sender details %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if strings.ToLower(tokenType) == "busy20" {
		tokenIssueFees.Busy20 = newFee
	} else if (strings.ToLower(tokenType) == "nft") || (strings.ToLower(tokenType) == "game") {
		tokenIssueFees.NFT = newFee
		tokenIssueFees.Game = newFee
	} else {
		return response, fmt.Errorf("Invalid token type , please send token type from [Busy20, NFT,Game]")
	}

	feesAsBytes, _ = json.Marshal(tokenIssueFees)
	err = ctx.GetStub().PutState("TokenIssueFees", feesAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while updating token issue fees on blockchain : %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	response.Message = "Token Issue Fee is set successfully"
	response.Success = true
	response.Data = tokenIssueFees
	logger.Info(response.Message)
	return response, nil
}

// IssueToken issue token in default wallet address of invoker
func (bt *Busy) IssueToken(ctx contractapi.TransactionContextInterface, tokenName string, symbol string, amount string, decimals uint64) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	bigAmount, isConverted := new(big.Int).SetString(amount, 10)
	if !isConverted {
		response.Message = "Error encountered converting amount"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if bigAmount.Cmp(bigZero) == 0 {
		response.Message = "Amount can not be zero"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	if decimals <= 0 && decimals > 18 {
		response.Message = "Decimals have to bein range of 1-18"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	commonName, _ := getCommonName(ctx)
	tokenFeeString, _ := GetTokenIssueFeeForTokenType(ctx, "busy20")
	issueTokenFee, _ := new(big.Int).SetString(tokenFeeString, 10)
	minusOne, _ := new(big.Int).SetString("-1", 10)
	defaultWalletAddress, err := getDefaultWalletAddress(ctx, commonName)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching user's default wallet: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	balance, err := getBalanceHelper(ctx, defaultWalletAddress, BUSY_COIN_SYMBOL)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching user's balance: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if balance.Cmp(issueTokenFee) == -1 {
		response.Message = "You do not have enough coins to issue new token"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
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
		response.Message = "NFT/Game token with same name already exists"
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	var token Token
	tokenAsBytes, err := ctx.GetStub().GetState(generateTokenStateAddress(symbol))
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching token details: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	tokenIdAsBytes, err := ctx.GetStub().GetState("latestTokenId")
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching latest token id: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	latestTokenID, _ := strconv.Atoi(string(tokenIdAsBytes))

	if tokenAsBytes == nil {
		var queryString string = fmt.Sprintf(`{
			"selector": {
				"docType": "token",
				"tokenName": "%s"
			 } 
		}`, tokenName)
		resultIterator, err := ctx.GetStub().GetQueryResult(queryString)
		if err != nil {
			response.Message = fmt.Sprintf("Error occured while fetching query data: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		defer resultIterator.Close()
		if resultIterator.HasNext() {
			response.Message = fmt.Sprintf("Token %s already exists", tokenName)
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}

		_ = ctx.GetStub().PutState("latestTokenId", []byte(strconv.Itoa(latestTokenID+1)))
		token := Token{
			DocType:     "token",
			ID:          uint64(latestTokenID + 1),
			TokenName:   tokenName,
			TokenSymbol: symbol,
			Admin:       commonName,
			TotalSupply: bigAmount.String(),
			Decimals:    decimals,
		}
		tokenAsBytes, _ = json.Marshal(token)
		err = ctx.GetStub().PutState(generateTokenStateAddress(symbol), tokenAsBytes)
		if err != nil {
			response.Message = fmt.Sprintf("Error occured while updating token on blockchain : %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}

		err = addTotalSupplyUTXO(ctx, symbol, bigAmount)
		if err != nil {
			response.Message = fmt.Sprintf("Error while creating total supply UTXO : %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		response.Data = token
	} else {
		// _ = json.Unmarshal(tokenAsBytes, &token)
		// if token.TokenName != tokenName {
		// 	response.Message = fmt.Sprintf("You must issue token with same name as before: %s", token.TokenName)
		// 	logger.Error(response.Message)
		// 	return response
		// }
		// bigTotalSupply, _ := new(big.Int).SetString(token.TotalSupply, 10)
		// token.TotalSupply = bigTotalSupply.Add(bigTotalSupply, bigAmount).String()
		// tokenAsBytes, _ = json.Marshal(token)
		// err = ctx.GetStub().PutState(symbol, tokenAsBytes)
		// if err != nil {
		// 	response.Message = fmt.Sprintf("Error while updating token on blockchain : %s", err.Error())
		// 	logger.Error(response.Message)
		// 	return response
		// }
		// response.Data = token
		response.Message = fmt.Sprintf("Token with symbol %s already exists", token.TokenSymbol)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// var queryString string = fmt.Sprintf(`{
	// 	"selector": {
	// 		"userId": "%s",
	// 		"docType": "wallet"
	// 	 }
	// }`, commonName)
	// resultIterator, err := ctx.GetStub().GetQueryResult(queryString)
	// if err != nil {
	// 	response.Message = fmt.Sprintf("Error while fetching admin wallet: %s", err.Error())
	// 	logger.Error(response.Message)
	// 	return response
	// }
	// defer resultIterator.Close()

	issuerAddress, _ := getDefaultWalletAddress(ctx, commonName)
	err = addUTXO(ctx, issuerAddress, bigAmount, symbol)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while generating utxo for new token: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	err = addUTXO(ctx, defaultWalletAddress, new(big.Int).Set(issueTokenFee).Mul(issueTokenFee, minusOne), BUSY_COIN_SYMBOL)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while burning fee for issue token: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = addTotalSupplyUTXO(ctx, BUSY_COIN_SYMBOL, new(big.Int).Set(issueTokenFee).Mul(minusOne, issueTokenFee))
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while burning issue token fee from total supply: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// var wallet Wallet
	// if resultIterator.HasNext() {
	// 	data, _ := resultIterator.Next()
	// 	_ = json.Unmarshal(data.Value, &wallet)
	// 	wallet.Balance += amount
	// 	walletAsBytes, _ := json.Marshal(token)
	// 	err = ctx.GetStub().PutState(symbol, walletAsBytes)
	// 	if err != nil {
	// 		response.Message = fmt.Sprintf("Error while updating wallet on blockchain : %s", err.Error())
	// 		logger.Error(response.Message)
	// 		return response
	// 	}
	// } else {
	// 	if err != nil {
	// 		response.Message = fmt.Sprintf("Can not issue token as wallet not found for user %s", commonName)
	// 		logger.Error(response.Message)
	// 		return response
	// 	}
	// }

	// Sending Balance Event
	balance, err = getBalanceHelper(ctx, defaultWalletAddress, BUSY_COIN_SYMBOL)
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching the balance: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	balanceData := []UserAddress{
		{
			Address: defaultWalletAddress,
			Token:   BUSY_COIN_SYMBOL,
		},
		{
			Address: defaultWalletAddress,
			Token:   symbol,
		},
	}
	balanceAsBytes, _ := json.Marshal(balanceData)
	err = ctx.GetStub().SetEvent(BALANCE_EVENT, balanceAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while Sending the Balance event: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	response.Message = fmt.Sprintf("Token %s has been successfully issued", symbol)
	response.Success = true
	logger.Info(response.Message)
	return response, nil
}

// Transfer transfer given amount from invoker's identity to specified identity
func (bt *Busy) Transfer(ctx contractapi.TransactionContextInterface, recipiant string, amount string, token string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	if amount == "0" {
		response.Message = "Zero amount can not be transferred"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// check if token exists
	exists, err := ifTokenExists(ctx, token)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching the details: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if !exists {
		response.Message = fmt.Sprintf("Symbol %s does not exist", token)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// check if wallet already exists
	walletAsBytes, err := ctx.GetStub().GetState(recipiant)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if walletAsBytes == nil {
		response.Message = fmt.Sprintf("Wallet %s does not exist", recipiant)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// Fetch current transfer fee
	transferFeesAsBytes, err := ctx.GetStub().GetState("transferFees")
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching transfer fee %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	bigTransferFee, _ := new(big.Int).SetString(string(transferFeesAsBytes), 10)

	bigAmount, _ := new(big.Int).SetString(amount, 10)

	if bigAmount == nil {
		response.Message = "Amount is invalid"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	// bigAmountWithTransferFee := bigAmount.Add(bigAmount, bigTransferFee)
	if token == "" {
		token = BUSY_COIN_SYMBOL
	}
	sender, _ := getCommonName(ctx)
	userAsBytes, err := ctx.GetStub().GetState(sender)
	if userAsBytes == nil {
		response.Message = fmt.Sprintf("User %s does not exist", sender)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching user %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	var user User
	_ = json.Unmarshal(userAsBytes, &user)

	if user.DefaultWallet == recipiant {
		response.Message = "It is not possible to transfer to the same address"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	isStakingAddress := strings.HasPrefix(recipiant, "staking-")
	if isStakingAddress {
		var wallet Wallet
		var queryString string = fmt.Sprintf(`{
			"selector": {
				"docType": "stakingAddr",
				"address": "%s"
			 } 
		}`, recipiant)
		resultIterator, err := ctx.GetStub().GetQueryResult(queryString)
		if err != nil {
			response.Message = fmt.Sprintf("Error occured while fetching user wallets: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		defer resultIterator.Close()

		if resultIterator.HasNext() {
			data, _ := resultIterator.Next()
			json.Unmarshal(data.Value, &wallet)
			if wallet.UserID != sender {
				response.Message = "It is not possible to make a transfer to the staking addresses"
				logger.Error(response.Message)
				return response, fmt.Errorf(response.Message)
			}
		} else {
			response.Message = "Staking address does not exist"
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
	}
	if sender == "busy_network" {
		bigTransferFee = bigZero
	}

	err = transferHelper(ctx, user.DefaultWallet, recipiant, bigAmount, token, bigTransferFee)
	if err != nil {
		//response.Message = fmt.Sprintf("Error occured while transfer: %s", err.Error())
		response.Message = "You do not have enough amount to transfer"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// minusOne, _ := new(big.Int).SetString("-1", 10)
	err = addTotalSupplyUTXO(ctx, BUSY_COIN_SYMBOL, bigTransferFee.Mul(minusOne, bigTransferFee))
	if err != nil {
		response.Message = fmt.Sprintf("Error while burning transfer fee: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// Sending Balance Event
	balanceData := []UserAddress{
		{
			Address: user.DefaultWallet,
			Token:   token,
		},
		{
			Address: user.DefaultWallet,
			Token:   BUSY_COIN_SYMBOL,
		},
		{
			Address: recipiant,
			Token:   token,
		},
	}
	balanceAsBytes, _ := json.Marshal(balanceData)
	err = ctx.GetStub().SetEvent(BALANCE_EVENT, balanceAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while Sending the Balance event: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	response.Message = "Transfer has been successfully accepted"
	logger.Info(response.Message)
	response.Success = true
	return response, nil
}

// GetTotalSupply get total supply of specified token
func (bt *Busy) GetTotalSupply(ctx contractapi.TransactionContextInterface, symbol string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	if symbol == "" {
		symbol = BUSY_COIN_SYMBOL
	}
	var token Token
	tokenAsBytes, err := ctx.GetStub().GetState(generateTokenStateAddress(symbol))
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching token details: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if tokenAsBytes == nil {
		response.Message = fmt.Sprintf("Symbol %s does not exist", symbol)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	_ = json.Unmarshal(tokenAsBytes, &token)

	totalSupply, _, err := pruneUTXOs(ctx, TOTAL_SUPPLY_KEY, symbol)
	if err != nil {
		response.Message = fmt.Sprintf("Error while prunning utxo: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	token.TotalSupply = totalSupply.String()

	response.Message = "Total supply has been successfully fetched"
	logger.Info(response.Message)
	response.Data = fmt.Sprintf("%s %s", token.TotalSupply, symbol)
	response.Success = true
	return response, nil
}

// Burn reduct balance from user wallet and reduce total supply
func (bt *Busy) Burn(ctx contractapi.TransactionContextInterface, address string, amount string, symbol string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	if amount == "0" {
		response.Message = "It is not possible to burn zero amount"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	exists, err := ifTokenExists(ctx, symbol)
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching token details: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if !exists {
		response.Message = fmt.Sprintf("Symbol %s does not exist", symbol)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// check if wallet already exists
	walletAsBytes, err := ctx.GetStub().GetState(address)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if walletAsBytes == nil {
		response.Message = fmt.Sprintf("Wallet %s does not exist", address)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	balance, err := getBalanceHelper(ctx, address, symbol)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching balance: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	bigAmount, _ := new(big.Int).SetString(amount, 10)
	if balance.Cmp(bigAmount) == -1 {
		response.Message = "There is not enough balance in the wallet"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	negetiveBigAmount, _ := new(big.Int).SetString("-"+amount, 10)
	mspid, _ := ctx.GetClientIdentity().GetMSPID()
	if mspid != "BusyMSP" {
		response.Message = "You are not allowed to burn BUSY coin"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	commonName, _ := getCommonName(ctx)
	if commonName != "busy_network" {
		response.Message = "You are not allowed to burn BUSY coin"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// var token Token
	// tokenAsBytes, err := ctx.GetStub().GetState(generateTokenStateAddress(symbol))
	// if tokenAsBytes == nil {
	// 	response.Message = fmt.Sprintf("Symbol %s does not exist", symbol)
	// 	logger.Error(response.Message)
	// 	return response, fmt.Errorf(response.Message)
	// }
	// if err != nil {
	// 	response.Message = fmt.Sprintf("Error occured while fetching token details: %s", err.Error())
	// 	logger.Error(response.Message)
	// 	return response, fmt.Errorf(response.Message)
	// }

	// _ = json.Unmarshal(tokenAsBytes, &token)
	// bigTotalSupply, _ := new(big.Int).SetString(token.TotalSupply, 10)
	// token.TotalSupply = bigTotalSupply.Sub(bigTotalSupply, bigAmount).String()
	// tokenAsBytes, _ = json.Marshal(token)
	err = addTotalSupplyUTXO(ctx, symbol, bigAmount)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while updating total supply: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	err = addUTXO(ctx, address, negetiveBigAmount, symbol)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while burning: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	defaultWalletAddress, err := getDefaultWalletAddress(ctx, commonName)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching user's default wallet: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = burnTxFee(ctx, defaultWalletAddress, BUSY_COIN_SYMBOL)
	if err != nil {
		response.Message = fmt.Sprintf("Error while burning tx fee: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if symbol == BUSY_COIN_SYMBOL {
		err = addTotalSupplyUTXO(ctx, BUSY_COIN_SYMBOL, negetiveBigAmount)
		if err != nil {
			response.Message = fmt.Sprintf("Error while updating total supply: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
	}
	balanceData := []UserAddress{
		{
			Address: address,
			Token:   symbol,
		},
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

	response.Message = "Burn has been successful"
	logger.Info(response.Message)
	response.Success = true
	return response, nil
}

// multibeneficiaryVestingV1 vesting v1
func (bt *Busy) MultibeneficiaryVestingV1(ctx contractapi.TransactionContextInterface, recipient string, amount string, numerator uint64, denominator uint64, releaseAt uint64) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	if amount == "0" {
		response.Message = "Zero amount can not be vested"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// check if wallet already exists
	walletAsBytes, err := ctx.GetStub().GetState(recipient)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if walletAsBytes == nil {
		response.Message = fmt.Sprintf("Wallet %s does not exist", recipient)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	now, _ := ctx.GetStub().GetTxTimestamp()
	mspid, _ := ctx.GetClientIdentity().GetMSPID()
	if mspid != "BusyMSP" {
		response.Message = "You are not allowed to create vesting"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	commonName, _ := getCommonName(ctx)
	if commonName != "busy_network" {
		response.Message = "You are not allowed to create vesting"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	bigAmount, _ := new(big.Int).SetString(amount, 10)
	if bigAmount.Cmp(bigZero) == 0 {
		response.Message = "Zero amount can not be vested"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	adminAddress, _ := getDefaultWalletAddress(ctx, commonName)
	balance, _ := getBalanceHelper(ctx, adminAddress, BUSY_COIN_SYMBOL)
	if balance.Cmp(bigAmount) == -1 {
		response.Message = "There is not enough balance in the wallet"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	lockedTokenAsBytes, _ := ctx.GetStub().GetState(fmt.Sprintf("vesting~%s", recipient))
	if lockedTokenAsBytes != nil {
		response.Message = fmt.Sprintf("Vesting for wallet %s already exists", recipient)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if releaseAt < uint64(now.Seconds) {
		response.Message = "Release time of vesting has to be in the future"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	totalAmount := new(big.Int).Set(bigAmount)
	currentVesting, err := calculatePercentage(bigAmount, numerator, denominator)
	if err != nil {
		response.Message = fmt.Sprintf("Error while Calculating Vesting percentage: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	txFee, err := getCurrentTxFee(ctx)
	bigTxFee, _ := new(big.Int).SetString(txFee, 10)
	if err != nil {
		response.Message = fmt.Sprintf("Error while getting tx fee: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = transferHelper(ctx, adminAddress, recipient, currentVesting, BUSY_COIN_SYMBOL, new(big.Int).Set(bigTxFee))
	if err != nil {
		//response.Message = fmt.Sprintf("Error occured while transfer: %s", err.Error())
		response.Message = "You do not have enough amount to transfer"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = addTotalSupplyUTXO(ctx, BUSY_COIN_SYMBOL, new(big.Int).Set(bigTxFee).Mul(minusOne, bigTxFee))
	if err != nil {
		response.Message = fmt.Sprintf("Error while burning transfer fee: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	lockedToken := LockedTokens{
		DocType:        "lockedToken",
		TotalAmount:    totalAmount.String(),
		ReleasedAmount: currentVesting.String(),
		StartedAt:      uint64(now.Seconds),
		ReleaseAt:      releaseAt,
	}
	lockedTokenAsBytes, _ = json.Marshal(lockedToken)
	err = ctx.GetStub().PutState(fmt.Sprintf("vesting~%s", recipient), lockedTokenAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while adding vesting schedule: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	balanceData := []UserAddress{
		{
			Address: recipient,
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
	response.Message = "Vesting has been scheduled successfully"
	logger.Info(response.Message)
	response.Success = true
	return response, nil
}

// multibeneficiaryVestingV2 vesting v2
func (bt *Busy) MultibeneficiaryVestingV2(ctx contractapi.TransactionContextInterface, recipient string, amount string, startAt uint64, releaseAt uint64) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	if amount == "0" {
		response.Message = "Zero amount can not be vested"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// check if wallet already exists
	walletAsBytes, err := ctx.GetStub().GetState(recipient)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if walletAsBytes == nil {
		response.Message = fmt.Sprintf("Wallet %s does not exist", recipient)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	now, _ := ctx.GetStub().GetTxTimestamp()
	mspid, _ := ctx.GetClientIdentity().GetMSPID()
	if mspid != "BusyMSP" {
		response.Message = "You are not allowed to create vesting"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	commonName, _ := getCommonName(ctx)
	if commonName != "busy_network" {
		response.Message = "You are not allowed to create vesting"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	bigAmount, _ := new(big.Int).SetString(amount, 10)
	if bigAmount.Cmp(bigZero) == 0 {
		response.Message = "Zero amount can not be vested"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	adminAddress, _ := getDefaultWalletAddress(ctx, commonName)
	balance, _ := getBalanceHelper(ctx, adminAddress, BUSY_COIN_SYMBOL)
	if balance.Cmp(bigAmount) == -1 {
		response.Message = "There is not enough balance in the wallet"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if releaseAt < startAt {
		response.Message = "Release time of vesting has to be greater then start time"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	lockedTokenAsBytes, _ := ctx.GetStub().GetState(fmt.Sprintf("vesting~%s", recipient))
	if lockedTokenAsBytes != nil {
		response.Message = fmt.Sprintf("Vesting for wallet %s already exists", recipient)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if releaseAt < uint64(now.Seconds) {
		response.Message = "Release time of vesting has to be in the future"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	if startAt < uint64(now.Seconds) {
		response.Message = "Start time of vesting has to be in the future"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	totalAmount := new(big.Int).Set(bigAmount)
	lockedToken := LockedTokens{
		DocType:        "lockedToken",
		TotalAmount:    totalAmount.String(),
		ReleasedAmount: "0",
		StartedAt:      startAt,
		ReleaseAt:      releaseAt,
	}
	lockedTokenAsBytes, _ = json.Marshal(lockedToken)
	err = ctx.GetStub().PutState(fmt.Sprintf("vesting~%s", recipient), lockedTokenAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while adding vesting schedule: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	sender, _ := getCommonName(ctx)
	defaultWalletAddress, err := getDefaultWalletAddress(ctx, sender)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = burnTxFee(ctx, defaultWalletAddress, BUSY_COIN_SYMBOL)
	if err != nil {
		response.Message = fmt.Sprintf("Error while burning transfer fee: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	balanceData := []UserAddress{
		{
			Address: recipient,
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
	response.Message = "Vesting has been scheduled successfully"
	logger.Info(response.Message)
	response.Success = true
	return response, nil
}

// GetLockedTokens get entry of vesting schedule for wallet address
func (bt *Busy) GetLockedTokens(ctx contractapi.TransactionContextInterface, address string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	lockedTokenAsBytes, err := ctx.GetStub().GetState(fmt.Sprintf("vesting~%s", address))
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while getting vesting details: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if lockedTokenAsBytes == nil {
		response.Message = fmt.Sprintf("Vesting entry does not exist for wallet %s", address)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	var lockedToken LockedTokens
	_ = json.Unmarshal(lockedTokenAsBytes, &lockedToken)

	response.Message = "Vesting has been successfully fetched"
	logger.Info(response.Message)
	response.Data = lockedToken
	response.Success = true
	return response, nil
}

// AttemptUnlock
func (bt *Busy) AttemptUnlock(ctx contractapi.TransactionContextInterface) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	commonName, _ := getCommonName(ctx)
	now, _ := ctx.GetStub().GetTxTimestamp()
	walletAddress, err := getDefaultWalletAddress(ctx, commonName)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	fee, _ := getCurrentTxFee(ctx)
	bigFee, _ := new(big.Int).SetString(fee, 10)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	balance, _ := getBalanceHelper(ctx, walletAddress, BUSY_COIN_SYMBOL)
	if bigFee.Cmp(balance) == 1 {
		response.Message = "There is not enough balance for tx fee in the wallet"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	lockedTokenAsBytes, err := ctx.GetStub().GetState(fmt.Sprintf("vesting~%s", walletAddress))
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while getting vesting details: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if lockedTokenAsBytes == nil {
		response.Message = fmt.Sprintf("Vesting entry does not exist for %s", walletAddress)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	var lockedToken LockedTokens
	_ = json.Unmarshal(lockedTokenAsBytes, &lockedToken)
	bigTotalAmount, _ := new(big.Int).SetString(lockedToken.TotalAmount, 10)
	bigReleasedAmount, _ := new(big.Int).SetString(lockedToken.ReleasedAmount, 10)
	bigStartedAt := new(big.Int).SetUint64(lockedToken.StartedAt)
	bigReleasedAt := new(big.Int).SetUint64(lockedToken.ReleaseAt)
	bigNow := new(big.Int).SetUint64(uint64(now.Seconds))

	if lockedToken.StartedAt > uint64(now.Seconds) {
		response.Message = "Vesting has not started yet"
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if lockedToken.ReleaseAt <= uint64(now.Seconds) {
		if lockedToken.TotalAmount == lockedToken.ReleasedAmount {
			response.Message = "There are no other coins to be claimed from this vesting"
			response.Success = false
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		amountToReleaseNow := bigTotalAmount.Sub(bigTotalAmount, bigReleasedAmount)
		lockedToken.ReleasedAmount = lockedToken.TotalAmount
		err = addUTXO(ctx, walletAddress, amountToReleaseNow, BUSY_COIN_SYMBOL)
		if err != nil {
			response.Message = fmt.Sprintf("Error occured while claiming: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		lockedTokenAsBytes, _ := json.Marshal(lockedToken)
		err = ctx.GetStub().PutState(fmt.Sprintf("vesting~%s", walletAddress), lockedTokenAsBytes)
		if err != nil {
			response.Message = fmt.Sprintf("Error occured while updating vesting schedule: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}

		// Burning the tx fee in attempt Unlock.
		err = burnTxFee(ctx, walletAddress, BUSY_COIN_SYMBOL)
		if err != nil {
			response.Message = fmt.Sprintf("Error while burning transfer fee: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		balanceData := []UserAddress{
			{
				Address: walletAddress,
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
		response.Message = "All tokens have been already unlocked"
		response.Success = true
		logger.Info(response.Message)
		return response, nil
	}
	releasableAmount := bigTotalAmount.Mul(bigNow.Sub(bigNow, bigStartedAt), bigTotalAmount).Div(bigTotalAmount, bigReleasedAt.Sub(bigReleasedAt, bigStartedAt))
	if releasableAmount.String() == "0" {
		response.Message = "There is nothing to release at this time"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if releasableAmount.Cmp(bigTotalAmount) == 1 {
		response.Message = "There is nothing to release at this time"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	releasableAmount = releasableAmount.Sub(releasableAmount, bigReleasedAmount)
	addUTXO(ctx, walletAddress, releasableAmount, BUSY_COIN_SYMBOL)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while claiming: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	bigReleasedAmount = bigReleasedAmount.Add(bigReleasedAmount, releasableAmount)
	lockedToken.ReleasedAmount = bigReleasedAmount.String()
	lockedTokenAsBytes, _ = json.Marshal(lockedToken)
	err = ctx.GetStub().PutState(fmt.Sprintf("vesting~%s", walletAddress), lockedTokenAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while updating vesting schedule: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	err = burnTxFee(ctx, walletAddress, BUSY_COIN_SYMBOL)
	if err != nil {
		response.Message = fmt.Sprintf("Error while burning transfer fee: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	balanceData := []UserAddress{
		{
			Address: walletAddress,
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

	response.Message = fmt.Sprintf("All tokens have been already unlocked")
	response.Success = true
	logger.Info(response.Message)
	return response, nil
}

func (bt *Busy) UpdateTransferFee(ctx contractapi.TransactionContextInterface, newTransferFee string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	mspid, _ := ctx.GetClientIdentity().GetMSPID()
	if mspid != "BusyMSP" {
		response.Message = "You are not allowed to set the transaction fee"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	commonName, _ := getCommonName(ctx)
	if commonName != "busy_network" {
		response.Message = "You are not allowed to set the transaction fee"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	err := ctx.GetStub().PutState("transferFees", []byte(newTransferFee))
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while updating transfer fee: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	sender, _ := getCommonName(ctx)
	defaultWalletAddress, err := getDefaultWalletAddress(ctx, sender)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	err = burnTxFee(ctx, defaultWalletAddress, BUSY_COIN_SYMBOL)
	if err != nil {
		response.Message = fmt.Sprintf("Error while burning transfer fee: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	response.Message = "Transfer fee has been successfully updated"
	response.Success = true
	response.Data = newTransferFee
	logger.Error(response.Message)
	return response, nil
}

func (bt *Busy) GetTokenDetails(ctx contractapi.TransactionContextInterface, tokenSymbol string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	tokenAsBytes, err := ctx.GetStub().GetState(generateTokenStateAddress(tokenSymbol))
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching token details: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if tokenAsBytes == nil {
		response.Message = fmt.Sprintf("Symbol %s does not exist", tokenSymbol)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	var token Token
	_ = json.Unmarshal(tokenAsBytes, &token)

	response.Message = "Token has been successfully fetched"
	response.Success = true
	response.Data = token
	logger.Info(response.Message)
	return response, nil
}

func (bt *Busy) GetStakingInfo(ctx contractapi.TransactionContextInterface, userID string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	userAsBytes, err := ctx.GetStub().GetState(userID)
	if userAsBytes == nil {
		response.Message = fmt.Sprintf("User %s does not exist", userID)
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching user from blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	userDetails := User{}
	if err := json.Unmarshal(userAsBytes, &userDetails); err != nil {
		response.Message = fmt.Sprintf("Error occured while retrieving sender details %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	var queryString string = fmt.Sprintf(`{
		"selector": {
			"userId": "%s",
			"docType": "stakingAddr"
		 } 
	}`, userID)
	resultIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	defer resultIterator.Close()

	var stakingAddr Wallet
	responseData := map[string]interface{}{}
	for resultIterator.HasNext() {
		tmpData := map[string]interface{}{}
		data, _ := resultIterator.Next()
		json.Unmarshal(data.Value, &stakingAddr)
		stakingInfo, _ := getStakingInfo(ctx, stakingAddr.Address)
		tmpData["claimed"] = stakingInfo.Claimed
		reward, err := countStakingReward(ctx, stakingInfo.StakingAddress)
		if err != nil {
			response.Message = fmt.Sprintf("Error occured while counting staking reward: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		tmpData["totalReward"] = reward.String()
		tmpData["creationTIme"] = stakingInfo.TimeStamp
		responseData[stakingAddr.Address] = tmpData
	}

	response.Message = "Staking details have been successfully fetched"
	response.Success = true
	response.Data = responseData
	logger.Info(response.Message)
	return response, nil
}

func (bt *Busy) Claim(ctx contractapi.TransactionContextInterface, stakingAddr string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	commonName, _ := getCommonName(ctx)
	fee, _ := getCurrentTxFee(ctx)
	bigFee, _ := new(big.Int).SetString(fee, 10)
	defaultWalletAddress, err := getDefaultWalletAddress(ctx, commonName)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	balance, _ := getBalanceHelper(ctx, defaultWalletAddress, BUSY_COIN_SYMBOL)
	if bigFee.Cmp(balance) == 1 {
		response.Message = "There is not enough balance for tx fee in the wallet"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	stakingAddrAsBytes, err := ctx.GetStub().GetState(stakingAddr)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching staking address: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if stakingAddrAsBytes == nil {
		response.Message = fmt.Sprintf("Staking address %s does not exist", stakingAddr)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	var stAddr Wallet
	json.Unmarshal(stakingAddrAsBytes, &stAddr)
	if stAddr.UserID != commonName {
		response.Message = "Ownership of the staking address has not been found"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	stakingReward, err := countStakingReward(ctx, stakingAddr)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while counting staking reward: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	logger.Infof("staking reward counted from countStakingReward func %s", stakingReward.String())

	stakingInfoAsBytes, err := ctx.GetStub().GetState(fmt.Sprintf("info~%s", stakingAddr))
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching staking details: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	var stakingInfo StakingInfo
	_ = json.Unmarshal(stakingInfoAsBytes, &stakingInfo)

	currentPhaseConfig, err := getPhaseConfig(ctx)
	if err != nil {
		response.Message = fmt.Sprintf("Error while initialising phase config: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// adminWalletAddress, _ := getDefaultWalletAddress(ctx, ADMIN_USER_ID)
	bigClaimedAmount, _ := new(big.Int).SetString(stakingInfo.Claimed, 10)
	logger.Info("Amout user claimed already fetching from staking info %s", bigClaimedAmount.String())
	claimableAmount := new(big.Int).Set(stakingReward).Sub(stakingReward, bigClaimedAmount)
	logger.Infof("claimable amout %s after deducting claimed amout %s from reward %s", claimableAmount.String(), bigClaimedAmount.String(), stakingReward.String())
	// err = transferHelper(ctx, adminWalletAddress, defaultWalletAddress, claimableAmount, BUSY_COIN_SYMBOL, bigZero)
	// if err != nil {
	// 	response.Message = fmt.Sprintf("Error while transfer from admin to default wallet: %s", err.Error())
	// 	logger.Error(response.Message)
	// 	return response
	// }
	claimableAmounAfterDeductingFee := new(big.Int).Set(claimableAmount).Sub(claimableAmount, bigFee)
	logger.Infof("claimable amout after deducting fee %s is %s", bigFee.String(), claimableAmounAfterDeductingFee.String())
	err = addUTXO(ctx, defaultWalletAddress, claimableAmounAfterDeductingFee, BUSY_COIN_SYMBOL)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while adding reward utxo: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	bigClaimedAmount = bigClaimedAmount.Add(bigClaimedAmount, claimableAmount)
	bigCurrentStakingAmount, _ := new(big.Int).SetString(stakingInfo.Amount, 10)
	bigCurrentStakingLimit, _ := new(big.Int).SetString(currentPhaseConfig.CurrentStakingLimit, 10)
	stakingInfo.Claimed = bigClaimedAmount.String()
	stakingInfo.Amount = currentPhaseConfig.CurrentStakingLimit
	stakingInfoAsBytes, _ = json.Marshal(stakingInfo)
	err = ctx.GetStub().PutState(fmt.Sprintf("info~%s", stakingAddr), stakingInfoAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while updating staking details: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	logger.Infof("staking reward before returning response ", stakingReward.String())
	stakingInfo.TotalReward = stakingReward.String()
	stakingInfo.Claimed = claimableAmount.String()

	amounOtherThenStakingLimit := bigCurrentStakingAmount.Sub(bigCurrentStakingAmount, bigCurrentStakingLimit)
	logger.Infof("amounOtherThenStakingLimit: %s", amounOtherThenStakingLimit.String())
	err = transferHelper(ctx, stakingAddr, defaultWalletAddress, amounOtherThenStakingLimit, BUSY_COIN_SYMBOL, bigZero)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while transferring from staking address to default wallet: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	err = addTotalSupplyUTXO(ctx, BUSY_COIN_SYMBOL, claimableAmounAfterDeductingFee)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while updating total supply: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	// err = burnTxFee(ctx, defaultWalletAddress, BUSY_COIN_SYMBOL)
	// if err != nil {
	// 	response.Message = fmt.Sprintf("Error while burning tx fee: %s", err.Error())
	// 	logger.Error(response.Message)
	// 	return response
	// }

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
	response.Message = "Claim has been successfully processed"
	response.Success = true
	response.Data = stakingInfo
	logger.Info(response.Message)
	return response, nil
}

func (bt *Busy) Unstake(ctx contractapi.TransactionContextInterface, stakingAddr string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	commonName, _ := getCommonName(ctx)
	fee, _ := getCurrentTxFee(ctx)
	bigFee, _ := new(big.Int).SetString(fee, 10)
	defaultWalletAddress, err := getDefaultWalletAddress(ctx, commonName)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching wallet %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	balance, _ := getBalanceHelper(ctx, defaultWalletAddress, BUSY_COIN_SYMBOL)
	if bigFee.Cmp(balance) == 1 {
		response.Message = "There is not enough balance for tx fee in the wallet"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	stakingAddrAsBytes, err := ctx.GetStub().GetState(stakingAddr)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching staking address: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if stakingAddrAsBytes == nil {
		response.Message = fmt.Sprintf("Staking address %s does not exist", stakingAddr)
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	var stAddr Wallet
	json.Unmarshal(stakingAddrAsBytes, &stAddr)
	if stAddr.UserID != commonName {
		response.Message = "Ownership of the staking address has not been found"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	stakingReward, err := countStakingReward(ctx, stakingAddr)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while counting staking reward: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	stakingInfoAsBytes, err := ctx.GetStub().GetState(fmt.Sprintf("info~%s", stakingAddr))
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching staking details: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	var stakingInfo StakingInfo
	_ = json.Unmarshal(stakingInfoAsBytes, &stakingInfo)

	// currentPhaseConfig, err := getPhaseConfig(ctx)
	// if err != nil {
	// 	response.Message = fmt.Sprintf("Error while getting phase config: %s", err.Error())
	// 	logger.Error(response.Message)
	// 	return response, fmt.Errorf(response.Message)
	// }
	// bigCurrentStakingLimit, _ := new(big.Int).SetString(phaseConfig.CurrentStakingLimit, 10)

	// adminWalletAddress, _ := getDefaultWalletAddress(ctx, ADMIN_USER_ID)
	bigClaimedAmount, _ := new(big.Int).SetString(stakingInfo.Claimed, 10)
	logger.Infof("Amount %s already claimed by %s", bigClaimedAmount.String(), stakingAddr)
	claimableAmount := new(big.Int).Set(stakingReward).Sub(stakingReward, bigClaimedAmount)
	logger.Infof("claimable amount after dedcuting claimed amount %s from total reward %s is %s", bigClaimedAmount.String(), stakingReward.String(), claimableAmount.String())
	// claimableAmount = claimableAmount.Add(claimableAmount, bigCurrentStakingLimit)
	bigStakingAmount, _ := new(big.Int).SetString(stakingInfo.Amount, 10)
	logger.Infof("staking amount for staking address %s is %s it is fetched from staking info", stakingAddr, bigStakingAmount.String())
	fmt.Println(bigZero)
	err = transferHelper(ctx, stakingAddr, defaultWalletAddress, bigStakingAmount, BUSY_COIN_SYMBOL, bigZero)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while transferring from staking address to default wallet: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	// err = transferHelper(ctx, adminWalletAddress, defaultWalletAddress, claimableAmount, BUSY_COIN_SYMBOL, bigZero)
	// if err != nil {
	// 	response.Message = fmt.Sprintf("Error while transfer from admin to default wallet: %s", err.Error())
	// 	logger.Error(response.Message)
	// 	return response
	// }
	claimableAmounAfterDeductingFee := new(big.Int).Set(claimableAmount).Sub(claimableAmount, bigFee)
	err = addUTXO(ctx, defaultWalletAddress, claimableAmounAfterDeductingFee, BUSY_COIN_SYMBOL)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while adding reward utxo: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	bigClaimedAmount = bigClaimedAmount.Add(bigClaimedAmount, claimableAmount)
	stakingInfo.Claimed = bigClaimedAmount.String()
	stakingInfo.Amount = "0"
	stakingInfoAsBytes, _ = json.Marshal(stakingInfo)
	err = ctx.GetStub().PutState(fmt.Sprintf("info~%s", stakingAddr), stakingInfoAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while updating staking details: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	stakingInfo.TotalReward = stakingReward.String()
	stakingInfo.Claimed = claimableAmount.String()

	err = ctx.GetStub().DelState(stakingAddr)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while deleting staking address: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	_, err = updateTotalStakingAddress(ctx, -1)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while updating number of total staking addresses: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// err = burnTxFee(ctx, defaultWalletAddress, BUSY_COIN_SYMBOL)
	// if err != nil {
	// 	response.Message = fmt.Sprintf("Error while burning tx fee: %s", err.Error())
	// 	logger.Error(response.Message)
	// 	return response
	// }
	err = addTotalSupplyUTXO(ctx, BUSY_COIN_SYMBOL, claimableAmounAfterDeductingFee)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while updating total supply: %s", err.Error())
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

	response.Message = "Claim and unstake been successfully processed"
	response.Success = true
	response.Data = stakingInfo
	logger.Info(response.Message)
	return response, nil
}

// GetCurrrentPhase config is to retrieve the current Phase config in Busy Chain
func (bt *Busy) GetCurrentPhase(ctx contractapi.TransactionContextInterface) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	currentPhaseConfig, err := getPhaseConfig(ctx)
	if err != nil {
		response.Message = fmt.Sprintf("Error while getting phase config: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	response.Success = true
	response.Message = "Current BusyChain phase has been successfully fetched"
	response.Data = currentPhaseConfig
	return response, nil
}

// GetCurrentFee config is to retrieve the current fees in Busy Chain
func (bt *Busy) GetCurrentFee(ctx contractapi.TransactionContextInterface) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	// Fetch current transfer fee
	transferFee, err := getCurrentTxFee(ctx)
	if err != nil {
		response.Message = fmt.Sprintf("Error occured while fetching transfer fee %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	response.Success = true
	response.Message = "Current transfer fee has been successfully fetched"
	response.Data = transferFee
	return response, nil
}

// AuthenticateUser config is to retrieve the current fees in Busy Chain
func (bt *Busy) AuthenticateUser(ctx contractapi.TransactionContextInterface, userID string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	// checking if user exists
	userAsBytes, err := ctx.GetStub().GetState(userID)
	if userAsBytes == nil {
		response.Message = fmt.Sprintf("User %s does not exist", userID)
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching user from blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	commonName, _ := getCommonName(ctx)
	if userID != commonName {
		response.Message = fmt.Sprintf("User %s does not match with certificate", userID)
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	response.Success = true
	response.Message = "User successfully Authenticated"
	response.Data = true
	return response, nil
}
