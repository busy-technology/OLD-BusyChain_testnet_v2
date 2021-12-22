
var atob = require('atob');

var AccessGroupsObject = {
    "busyadmin": [
        /* Section 2 - Account */ "register", "login", "queryWalletBalances", "userWallets", "recoverUser", "resetPassword", "defaultWallets",
        /* Section 3 - Staking */ "createStakingAddress", "stakingInfo", "claim", "unstake", "stakingAddresses",
        /* Section 4 - BusyChain */ "transfer", "getTotalSupply", "getBlocks", "transactions", "currentPhase", "transactionFees", "messagingFees", "tokenIssueFees",
        /* Section 6 - Voting Governance */ "createPool", "createVote", "poolConfig", "poolHistory", "queryPool",
        /* Section 7 - Messaging */ "sendMessage",
        /* Section 8 - Tokens BUSY20 */ "issue", "issuedTokens",
        /* Section 8 - Tokens NFT+Game */ "nft/mintToken", "nft/mintBatch", "nft/burnBatch", "nft/transfer", "nft/transferBatch", "nft/checkApproval", "nft/setApproval", "nft/balance", "nft/balanceBatch", "nft/mintedTokens", "nft/tokenInfo", "nft/updateTokenMetaData",
        /* Section 8 - Tokens BUSYNFT */ "busynft/mint", "busynft/transfer", "busynft/getCurrentOwner", "busynft/updateTokenMetaData", "busynft/mintedTokens",
        /* Section 9 - ADMIN Functions*/ "swap", "burn", "destroyPool",
        /* Section 9 - ADMIN Configurations*/ "poolConfig", "updateTransferFees", "updateMessagingFees", "updateTokenIssueFees",
        /* Section 9 - ADMIN - Vesting */ "lockedVestingInfo", "attemptUnlock", "vestingV1", "vestingV2"
    ],

    "busytestnet": [
        /* Section 2 - Account */ "register", "login", "queryWalletBalances", "userWallets", "recoverUser", "resetPassword", "defaultWallets",
        /* Section 3 - Staking */ "createStakingAddress", "stakingInfo", "claim", "unstake", "stakingAddresses",
        /* Section 4 - BusyChain */ "transfer", "getTotalSupply", "getBlocks", "transactions", "currentPhase", "transactionFees", "messagingFees", "tokenIssueFees",
        /* Section 6 - Voting Governance */ "createPool", "createVote", "poolConfig", "poolHistory", "queryPool",
        /* Section 7 - Messaging */ "sendMessage",
        /* Section 8 - Tokens BUSY20 */ "issue", "issuedTokens",
        /* Section 8 - Tokens NFT+Game */ "nft/mintToken", "nft/mintBatch", "nft/burnBatch", "nft/transfer", "nft/transferBatch", "nft/checkApproval", "nft/setApproval", "nft/balance", "nft/balanceBatch", "nft/mintedTokens", "nft/tokenInfo", "nft/updateTokenMetaData",
        /* Section 8 - Tokens BUSYNFT */ "busynft/mint", "busynft/transfer", "busynft/getCurrentOwner", "busynft/updateTokenMetaData", "busynft/mintedTokens"
    ],

    "busyuser": [
        /* Section 2 - Account */ "register", "userWallets", "defaultWallets",
        /* Section 3 - Staking */ "stakingInfo", "stakingAddresses",
        /* Section 4 - BusyChain */ "transfer", "getTotalSupply", "getBlocks", "transactions", "currentPhase", "transactionFees", "messagingFees", "tokenIssueFees",
        /* Section 6 - Voting Governance */ "poolHistory", "queryPool",
        /* Section 7 - Messaging */
        /* Section 8 - Tokens BUSY20 */
        /* Section 8 - Tokens NFT+Game */
        /* Section 8 - Tokens BUSYNFT */
        /* Section 8 - Tokens */ "issuedTokens",
        /* Section 9 - ADMIN */
        /* Section 9 - ADMIN - Vesting */ "lockedVestingInfo"
    ],

    "busywallet": [
        /* Section 2 - Account */ "register", "login", "queryWalletBalances", "userWallets", "recoverUser", "resetPassword", "defaultWallets",
        /* Section 3 - Staking */ "createStakingAddress", "stakingInfo", "claim", "unstake", "stakingAddresses",
        /* Section 4 - BusyChain */ "transfer", "getTotalSupply", "getBlocks", "transactions", "currentPhase", "transactionFees", "messagingFees", "tokenIssueFees",
        /* Section 6 - Voting Governance */ "poolHistory", "queryPool",
        /* Section 7 - Messaging */
        /* Section 8 - Tokens BUSY20 */ "issue", "issuedTokens",
        /* Section 8 - Tokens NFT+Game */ "nft/mintToken", "nft/mintBatch", "nft/burnBatch", "nft/transfer", "nft/transferBatch", "nft/checkApproval", "nft/setApproval", "nft/balance", "nft/balanceBatch", "nft/mintedTokens", "nft/tokenInfo", "nft/updateTokenMetaData",
        /* Section 8 - Tokens BUSYNFT */ "busynft/mint", "busynft/transfer", "busynft/getCurrentOwner", "busynft/updateTokenMetaData", "busynft/mintedTokens",
                /* Section 9 - ADMIN */
                /* Section 9 - ADMIN - Vesting */
    ]
};

module.exports = (token, functionName) => {

    if (token == null) {
        return {
            authorized: false,
            errorMsg: "Authorization header not preset",
            statusCode: 401
        };
    }
    ;

    payload = parseJwt(token);
    console.log("Checking the authorization for ", functionName);


    if (payload.domainname in AccessGroupsObject) {
        let functionList = AccessGroupsObject[payload.domainname];

        if (functionList.includes(functionName)) {
            return {
                authorized: true,
                statusCode: 200
            };
        } else {
            return {
                authorized: false,
                errorMsg: "Forbidden",
                statusCode: 403
            };
        }
    } else {
        return {
            authorized: false,
            errorMsg: "Forbidden",
            statusCode: 403
        };
    }
    ;
};


function parseJwt(token) {
    var base64Url = token.split('.')[1];
    var base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    var jsonPayload = decodeURIComponent(atob(base64).split('').map(function (c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));

    return JSON.parse(jsonPayload);
}