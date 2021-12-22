const errors = require("restify-errors"),
        trim_req = require("./libs/request/trim"),
        controller = require("./controllers"),
        middleware = require("./middlewares"),
        auth = require("./middlewares/auth");

/**
 * List of routes
 * @param RestifyServer server
 */

module.exports = (server) => {
// trim request parameter
    server.use(trim_req);
    server.post(
            "/auth/generate-token",
            middleware.auth.generateToken,
            controller.auth.generateToken
            );
    /*server.post(
     "/auth/generate-token-admin",
     middleware.auth.generateToken,
     controller.auth.generateTokenAdmin
     );*/
    /**
     * @description User registration
     * @date july-06-2021
     * @author Raj
     */


    server.post(
            "/register",
            auth,
            middleware.utility.required(["userId", "password", "confirmPassword"]),
            middleware.utility.number(["mobile"]),
            middleware.utility.userId(["userId"]),
            middleware.utility.isName(["firstName"]),
            middleware.utility.isName(["lastName"]),
            middleware.utility.isCountry(["country"]),
            middleware.utility.isPassword(["password"]),
            middleware.utility.isPassword(["confirmPassword"]),
            middleware.utility.isEmail(["email"]),
            controller.users.register
            );
    server.post(
            "/login",
            auth,
            middleware.utility.required(["userId", "password"]),
            controller.users.login
            );
    server.post(
            "/createStakingAddress",
            auth,
            middleware.utility.required(["userId", "credentials", "type"]),
            controller.users.wallet
            );
    server.post(
            "/attemptUnlock",
            auth,
            middleware.utility.required(["walletId", "credentials"]),
            controller.users.attemptUnlock
            );
    server.post(
            "/swap",
            auth,
            middleware.utility.required(["recipiant", "amount", "token"]),
            middleware.utility.isAmount(["amount"]),
            controller.users.swap
            );
    server.post(
            "/transfer",
            auth,
            middleware.utility.required(["sender", "credentials", "recipiant", "amount", "token"]),
            middleware.utility.isAmount(["amount"]),
            controller.users.transfer
            );
    server.get(
            "/getBlocks",
            auth,
            controller.users.getBlocks
            );
    server.post(
            "/claim",
            auth,
            middleware.utility.required(["stakingAddr", "credentials"]),
            controller.users.claim
            );
    server.post(
            "/unstake",
            auth,
            middleware.utility.required(["stakingAddr", "credentials"]),
            controller.users.unstake
            );
    server.post(
            "/issue",
            auth,
            middleware.utility.required(["walletId", "credentials", "tokenName", "symbol", "amount", "decimals"]),
            middleware.utility.isAmount(["amount"]),
            middleware.utility.isAmount(["decimals"]),
            middleware.utility.isAlphaNumeric(["tokenName"]),
            middleware.utility.isAlphaNumeric(["symbol"]),
            controller.users.issue
            );
    server.post(
            "/getTotalSupply",
            auth,
            middleware.utility.required(["symbol"]),
            controller.users.totalSupply
            );
    server.post(
            "/updateTransferFees",
            auth,
            middleware.utility.required(["newTransferFee"]),
            middleware.utility.isAmount(["newTransferFee"]),
            controller.users.transferFee
            );
    server.post(
            "/burn",
            auth,
            middleware.utility.required(["walletId", "amount", "token"]),
            middleware.utility.isAmount(["amount"]),
            controller.users.burn
            );
    server.post(
            "/vestingV1",
            auth,
            middleware.utility.required([
                "recipient",
                "amount",
                "numerator",
                "denominator",
                "releaseAt"
            ]),
            middleware.utility.isAmount(["amount"]),
            middleware.utility.isNumeratorDenominator(["numerator", "denominator"]),
            controller.users.vesting1
            );
    server.post(
            "/vestingV2",
            auth,
            middleware.utility.required(["recipient", "amount", "startAt", "releaseAt"]),
            middleware.utility.isAmount(["amount"]),
            middleware.utility.isTime(["startAt", "releaseAt"]),
            controller.users.vesting2
            );
    server.post(
            "/lockedVestingInfo",
            auth,
            middleware.utility.required(["walletId"]),
            controller.users.lockedVestingInfo
            );
// server.post(
//   "/queryWallet",
//   middleware.utility.required(["userId", "credentials"]),
//   controller.users.queryWallet
// );


    server.post(
            "/queryWalletBalances",
            auth,
            middleware.utility.required(["walletId"]),
            controller.users.queryWalletAdmin
            );
    server.get(
            "/stakingAddresses",
            auth,
            controller.users.fetchWallets
            );
    server.get(
            "/defaultWallets",
            auth,
            controller.users.defaultWallets
            );
    server.get(
            "/issuedTokens",
            auth,
            controller.users.issuedTokens
            );
    server.get(
            "/currentPhase",
            auth,
            controller.users.getCurrentPhase
            );
    server.get(
            "/transactionFees",
            auth,
            controller.users.getCurrentFee
            );
    server.get(
            "/transactions",
            auth,
            controller.users.transactions
            );
    server.post(
            "/stakingInfo",
            auth,
            controller.users.stakingInfo
            );
    server.post(
            "/recoverUser",
            auth,
            middleware.utility.required(["userId", "mnemonic"]),
            controller.users.recoverUser
            );
    server.post(
            "/resetPassword",
            auth,
            middleware.utility.required(["userId", "newPassword", "credentials"]),
            middleware.utility.isPassword(["newPassword"]),
            controller.users.resetPassword
            );
// server.post(
//   "/addAdmin",
//   middleware.utility.required(["credentials"]),
//   auth,
//   controller.users.addAdmin
// );

// Add in user wallets

    server.post(
            "/userWallets",
            auth,
            middleware.utility.required(["userId"]),
            controller.users.userWallets
            );
    server.post(
            "/sendMessage",
            auth,
            middleware.utility.required(["sender", "recipient"]),
            controller.users.sendMessage
            );
// endpoint for creating pool
    server.post(
            "/createPool",
            auth,
            middleware.utility.required([
                "walletId",
                "credentials",
                "poolName",
                "poolDescription"
            ]),
            middleware.utility.isPoolName(["poolName"]),
            middleware.utility.isPoolDescription(["poolDescription"]),
            controller.users.createPool
            );
// endpoint for pool Config
    server.get(
            "/poolConfig",
            auth,
            controller.users.getPoolConfig
            );
// endpoint for pool Config
    server.post(
            "/poolConfig",
            auth,
            middleware.utility.required([
                "minimumCoins",
                "poolFee",
                "votingPeriod",
                "votingStartTime"
            ]),
            middleware.utility.isAmount(["poolFee"]),
            middleware.utility.isAmount(["minimumCoins"]),
            middleware.utility.isTime(["votingPeriod"]),
            middleware.utility.isTime(["votingStartTime"]),
            controller.users.updatePoolConfig
            );
// endpoint for creating pool
    server.get(
            "/queryPool",
            auth,
            controller.users.queryPool
            );
// endpoint for pool history
    server.get(
            "/poolHistory",
            auth,
            controller.users.poolHistory
            );
// endpoint for minting tokens to a new account
    server.post(
            "/nft/mintToken",
            auth,
            middleware.utility.required([
                "walletId",
                "tokenSymbol",
                "totalSupply",
                "credentials",
                "metaData"
            ]),
            middleware.utility.isNumeric(["totalSupply"]),
            controller.users.mintToken
            );
// endpoint for minting batch of tokens to a new account
    server.post(
            "/nft/mintBatch",
            auth,
            middleware.utility.required([
                "walletId",
                "tokenSymbols",
                "totalSupplies",
                "credentials",
                "metaDatas"
            ]),
            middleware.utility.isArray(["tokenSymbols", "totalSupplies"]),
            middleware.utility.isAmounts(["totalSupplies"]),
            controller.users.mintBatch
            );
// endpoint for burn batch of tokens to a new account
    server.post(
            "/nft/burnBatch",
            auth,
            middleware.utility.required([
                "walletId",
                "tokenSymbols",
                "amounts"
            ]),
            middleware.utility.isArray(["tokenSymbols", "amounts"]),
            middleware.utility.isAmounts(["amounts"]),
            controller.users.burnBatch
            );
// endpoint for transfer of tokens from sender to receipient
    server.post(
            "/nft/transfer",
            auth,
            middleware.utility.required([
                "account",
                "operator",
                "tokenSymbol",
                "amount",
                "recipient",
                "credentials"
            ]),
            middleware.utility.isNumeric(["amount"]),
            controller.users.nftTransfer,
            );
// endpoint for transfer batch of tokens from sender to receipient
    server.post(
            "/nft/transferBatch",
            auth,
            middleware.utility.required([
                "account",
                "operator",
                "tokenSymbols",
                "amounts",
                "recipient",
                "credentials"
            ]),
            middleware.utility.isArray(["tokenSymbols", "amounts"]),
            controller.users.nftTransferBatch,
            middleware.utility.isAmounts(["amounts"]),
            );
// endpoint for get Approval for the tokens of an account
    server.post(
            "/nft/checkApproval",
            auth,
            middleware.utility.required([
                "walletId",
                "credentials",
                "operator"
            ]),
            controller.users.checkApproval,
            );
// endpoint for set Approval for the tokens of an account
    server.post(
            "/nft/setApproval",
            auth,
            middleware.utility.required([
                "walletId",
                "credentials",
                "operator"
            ]),
            controller.users.setApproval,
            );
// endpoint for balance of token in the caller account
    server.post(
            "/nft/balance",
            auth,
            middleware.utility.required([
                "walletId",
                "tokenSymbol"
            ]),
            controller.users.nftBalance,
            );
// endpoint for balance of differnt tokens in the caller account
    server.post(
            "/nft/balanceBatch",
            auth,
            middleware.utility.required([
                "walletIds",
                "tokenSymbols"
            ]),
            middleware.utility.isArray(["tokenSymbols"]),
            controller.users.nftBalanceBatch,
            );
// endpoint for getting metadata of the token
    server.get(
            "/nft/tokenInfo",
            auth,
            controller.users.nftTokenInfo,
            );
// endpoint for getting all the minted Tokens
    server.get(
            "/nft/mintedTokens",
            auth,
            controller.users.mintedTokens,
            );
// endpoint for Updating token Metadata
    server.post(
            "/nft/updateTokenMetaData",
            auth,
            middleware.utility.required([
                "walletId",
                "tokenSymbol",
                "credentials",
                "metaData"
            ]),
            middleware.utility.isNumeric(["totalSupply"]),
            controller.users.updateTokenMetaData
            );
// endpoint for minting busy nft tokens to a new account
    server.post(
            "/busynft/mint",
            auth,
            middleware.utility.required([
                "walletId",
                "nftName",
                "credentials",
                "metaData"
            ]),
            controller.users.busyNftmint
            );
// endpoint for transfer of busy nft tokens from sender to receipient
    server.post(
            "/busynft/transfer",
            auth,
            middleware.utility.required([
                "sender",
                "nftName",
                "recipient",
                "credentials"
            ]),
            controller.users.busyNftTransfer,
            );
// endpoint for transfer of tokens from sender to receipient
    server.get(
            "/busynft/mintedTokens",
            auth,
            controller.users.getSpecialMintedTokens,
            );

// endpoint for getting special minted tokens
    server.get(
            "/busynft/getCurrentOwner",
            auth,
            middleware.utility.required([
                "nftName"
            ]),
            controller.users.getCurrentOwner,
            );

// endpoint for updating nft tokens
    server.post(
            "/busynft/updateTokenMetaData",
            auth,
            middleware.utility.required([
                "walletId",
                "nftName",
                "credentials",
                "metaData"
            ]),
            middleware.utility.isNumeric(["totalSupply"]),
            controller.users.updateSpecialMetaData
            );
// endpoint for creating vote
    server.post(
            "/createVote",
            auth,
            middleware.utility.required([
                "walletId",
                "credentials",
                "votingAddress",
                "amount",
                "voteType"
            ]),
            middleware.utility.isAmount(["amount"]),
            middleware.utility.voteType(["voteType"]),
            controller.users.createVote
            );
// endpoint for destroying the pool
    server.post(
            "/destroyPool",
            auth,
            controller.users.destroyPool
            );


    server.get(
            "/tokenIssueFees",
            auth,
            controller.users.getTokenIssueFee
            );
    server.get(
            "/messagingFees",
            auth,
            controller.users.getMessagingFee
            );
    server.post(
            "/updateTokenIssueFees",
            auth,
            middleware.utility.required(["newFee", "tokenType"]),
            middleware.utility.isAmount(["newFee"]),
            controller.users.updateTokenIssueFee
            );
    server.post(
            "/updateMessagingFees",
            auth,
            middleware.utility.required(["newFee"]),
            middleware.utility.isAmount(["newFee"]),
            controller.users.updateMessagingFee
            );
};