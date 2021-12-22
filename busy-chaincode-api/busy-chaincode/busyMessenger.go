package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type LastMessage struct {
	MessageTime time.Time
	Sender      string
	Recipient   string
}

// BusyMessenger contract
type BusyMessenger struct {
	contractapi.Contract
}

// MessageInfo
type MessageStore struct {
	Sender    map[string]int
	Recipient map[string]int
}

// CreateUser creates new user on busy blockchain
func (bm *BusyMessenger) SendMessage(ctx contractapi.TransactionContextInterface, recipient string, token string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	sender, _ := getCommonName(ctx)
	logger.Info("Recieved a message from", sender, "to", recipient)

	// getting the default config for messaging functionality
	configAsBytes, err := ctx.GetStub().GetState("MessageConfig")
	if err != nil {
		response.Message = fmt.Sprintf("Error while getting confing state: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	var config MessageConfig
	if err = json.Unmarshal(configAsBytes, &config); err != nil {
		response.Message = fmt.Sprintf("Error while unmarshalling the confing state: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// getting the last Message(time, sender and reciever) State for a single user
	lastMessageAsBytes, err := ctx.GetStub().GetState(getLastMessageKey(sender))
	if err != nil {
		response.Message = fmt.Sprintf("Error while getting last Message state: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if lastMessageAsBytes != nil {
		var lastMessage LastMessage
		_ = json.Unmarshal(lastMessageAsBytes, &lastMessage)
		if time.Now().Sub(lastMessage.MessageTime) < config.MessageInterval {
			response.Message = fmt.Sprintf("Please wait for 5 seconds before sending the next message")
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
	}
	//updating the last Message
	lastMessage := LastMessage{
		MessageTime: time.Now(),
		Sender:      sender,
		Recipient:   recipient,
	}
	lastMessageAsBytes, _ = json.Marshal(lastMessage)
	err = ctx.GetStub().PutState(getLastMessageKey(sender), lastMessageAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	if sender == recipient {
		//response.Message = fmt.Sprintf("message cannot be sent to the same userId: %s", sender)
		response.Message = "You cannot send the message to yourself"
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	senderAsBytes, err := ctx.GetStub().GetState(sender)
	if senderAsBytes == nil {
		response.Message = fmt.Sprintf("Sender with common name %s does not exists", sender)
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching user from blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	senderDetails := User{}
	if err := json.Unmarshal(senderAsBytes, &senderDetails); err != nil {
		response.Message = fmt.Sprintf("Error while retrieving the sender details %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	recipientAsBytes, err := ctx.GetStub().GetState(recipient)
	if recipientAsBytes == nil {
		response.Message = fmt.Sprintf("Recipient with common name %s does not exists", recipient)
		logger.Info(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	if err != nil {
		response.Message = fmt.Sprintf("Error while fetching user from blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	recipientDetails := User{}
	if err := json.Unmarshal(recipientAsBytes, &recipientDetails); err != nil {
		response.Message = fmt.Sprintf("Error while retrieving the recipient details %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	val, ok := senderDetails.MessageCoins[recipientDetails.DefaultWallet]

	var messagestore MessageStore
	// using MessageStore
	if ok && val > 0 {
		logger.Info("Using the message store")
		if err := AddCoins(ctx, recipientDetails.DefaultWallet, config.BigBusyCoins, token); err != nil {
			response.Message = fmt.Sprintf("Error while Adding coins to the recipient default wallet %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		if val == config.BusyCoin {
			// deleting the key from map
			delete(senderDetails.MessageCoins, recipientDetails.DefaultWallet)
		} else {
			senderDetails.MessageCoins[recipientDetails.DefaultWallet] = val - config.BusyCoin
		}
		senderDetails.MessageCoins["totalCoins"] -= config.BusyCoin
		senderAsBytes, err = json.Marshal(senderDetails)
		if err != nil {
			response.Message = fmt.Sprintf("Error while Marshalling the senderdetails %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		err = ctx.GetStub().PutState(sender, senderAsBytes)
		if err != nil {
			response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		messagestore.Sender = senderDetails.MessageCoins
		messagestore.Recipient = recipientDetails.MessageCoins
	} else {
		logger.Info("using default wallet")

		balance, _ := getBalanceHelper(ctx, senderDetails.DefaultWallet, token)
		amountInt, _ := new(big.Int).SetString(config.BigBusyCoins, 10)
		if balance.Cmp(amountInt) == -1 {
			//response.Message = fmt.Sprintf("User: %s does not have enough coins to Send Message", sender)
			response.Message = "You do not have enough coins to send a message"
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}

		if err := RemoveCoins(ctx, senderDetails.DefaultWallet, config.BigBusyCoins, token); err != nil {
			response.Message = fmt.Sprintf("Error while Adding coins to the recipient default wallet %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		if val, ok := recipientDetails.MessageCoins[senderDetails.DefaultWallet]; ok {
			recipientDetails.MessageCoins[senderDetails.DefaultWallet] = val + config.BusyCoin
		} else {
			recipientDetails.MessageCoins[senderDetails.DefaultWallet] = config.BusyCoin
		}
		recipientDetails.MessageCoins["totalCoins"] += config.BusyCoin

		recipientAsBytes, err = json.Marshal(recipientDetails)
		if err != nil {
			response.Message = fmt.Sprintf("Error while Marshalling the recipientDetails %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		err = ctx.GetStub().PutState(recipient, recipientAsBytes)
		if err != nil {
			response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
			logger.Error(response.Message)
			return response, fmt.Errorf(response.Message)
		}
		messagestore.Sender = senderDetails.MessageCoins
		messagestore.Recipient = recipientDetails.MessageCoins
	}
	balanceData := []UserAddress{
		{
			Address: senderDetails.DefaultWallet,
			Token:   BUSY_COIN_SYMBOL,
		},
		{
			Address: recipientDetails.DefaultWallet,
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
	response.Data = messagestore
	response.Message = "Message has been sent successfully"
	response.Success = true
	return response, nil
}

//function to update messaging fee
func (bm *BusyMessenger) UpdateMessagingFee(ctx contractapi.TransactionContextInterface, newFee string) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	//check whether admin or not

	commonName, _ := getCommonName(ctx)
	if commonName != "busy_network" {
		response.Message = "Only admin can update the messaging fee"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// getting the default config for messaging functionality
	configAsBytes, err := ctx.GetStub().GetState("MessageConfig")
	if err != nil {
		response.Message = fmt.Sprintf("Error while getting confing state: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	var config MessageConfig
	if err = json.Unmarshal(configAsBytes, &config); err != nil {
		response.Message = fmt.Sprintf("Error while unmarshalling the confing state: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	//updating the messaging fee

	//validate the newFee using BigInt
	config.BigBusyCoins = newFee

	configAsBytes, _ = json.Marshal(config)
	err = ctx.GetStub().PutState("MessageConfig", configAsBytes)
	if err != nil {
		response.Message = fmt.Sprintf("Error while updating state in blockchain: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	response.Data = config.BigBusyCoins
	response.Message = "Messaging fee has been updated successfully"
	response.Success = true
	return response, nil
}

func (bm *BusyMessenger) GetMessagingFee(ctx contractapi.TransactionContextInterface) (*Response, error) {
	response := &Response{
		TxID:    ctx.GetStub().GetTxID(),
		Success: false,
		Message: "",
		Data:    nil,
	}

	//check whether admin or not

	commonName, _ := getCommonName(ctx)
	if commonName != "busy_network" {
		response.Message = "Only admin can update the messaging fee"
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	// getting the default config for messaging functionality
	configAsBytes, err := ctx.GetStub().GetState("MessageConfig")
	if err != nil {
		response.Message = fmt.Sprintf("Error while getting confing state: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}
	var config MessageConfig
	if err = json.Unmarshal(configAsBytes, &config); err != nil {
		response.Message = fmt.Sprintf("Error while unmarshalling the confing state: %s", err.Error())
		logger.Error(response.Message)
		return response, fmt.Errorf(response.Message)
	}

	//updating the messaging fee

	//validate the newFee using BigInt

	response.Data = config.BigBusyCoins
	response.Message = "Current messaging fee has been successfully fetched"
	response.Success = true
	return response, nil
}

// RemoveCoins is to move coins from default wallet to message store
func RemoveCoins(ctx contractapi.TransactionContextInterface, address string, coins string, token string) error {
	minusOne, _ := new(big.Int).SetString("-1", 10)
	bigTxFee, _ := new(big.Int).SetString(coins, 10)

	utxo := UTXO{
		DocType: "utxo",
		Address: address,
		Amount:  bigTxFee.Mul(bigTxFee, minusOne).String(),
		Token:   BUSY_COIN_SYMBOL,
	}
	utxoAsBytes, _ := json.Marshal(utxo)
	err := ctx.GetStub().PutState(fmt.Sprintf("message~%s~%s~%s", ctx.GetStub().GetTxID(), address, BUSY_COIN_SYMBOL), utxoAsBytes)
	if err != nil {
		return err
	}
	return nil
}

// RemoveCoins is to move coins from default wallet to message store
func AddCoins(ctx contractapi.TransactionContextInterface, address string, coins string, token string) error {
	plusOne, _ := new(big.Int).SetString("1", 10)
	bigTxFee, _ := new(big.Int).SetString(coins, 10)

	utxo := UTXO{
		DocType: "utxo",
		Address: address,
		Amount:  bigTxFee.Mul(bigTxFee, plusOne).String(),
		Token:   BUSY_COIN_SYMBOL,
	}
	utxoAsBytes, _ := json.Marshal(utxo)
	err := ctx.GetStub().PutState(fmt.Sprintf("message~%s~%s~%s", ctx.GetStub().GetTxID(), address, BUSY_COIN_SYMBOL), utxoAsBytes)
	if err != nil {
		return err
	}
	return nil
}
