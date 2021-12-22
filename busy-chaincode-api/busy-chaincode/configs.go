package main

type TokenIssueFee struct {
	// BusyCoins to deduct
	Busy20 string `json:"busy20"`
	NFT    string `json:"nft"`
	Game   string `json:"game"`
}
