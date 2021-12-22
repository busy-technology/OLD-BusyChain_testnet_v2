const {
    Gateway,
    Wallets
} = require("fabric-network");
const protobuf = require('protobufjs');
const blocks = require("../app/models/blocks");
const BlockDecoder = require("fabric-client/lib/BlockDecoder");
const fs = require("fs");
const path = require("path");


exports.FabricInvokeBlocks = async (
    channelName,
    contractName,
    functionName,
    userId,
    userKey,
    txId,
) => {
    try {
        // load the network configuration
        const ccpPath = path.resolve(
            __dirname,
            "connection-profile",
            "connection-busy.json"
        );
        const ccp = JSON.parse(fs.readFileSync(ccpPath, "utf8"));
        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(
            process.cwd(),
            "blockchain",
            "network",
            "wallet"
        );
        // const walletPath = path.resolve(__dirname, '..', '..', 'network', 'wallet')
        const wallet = await Wallets.newFileSystemWallet(walletPath);


        await wallet.put(userId, userKey);

        // Check to see if we've already enrolled the user.
        // const identity = await wallet.exists(arguements.akcessId);
        // const identity = await wallet.get(userId);
        // if (!identity) {
        //   console.log("An identity for the user does not exist in the wallet");
        //   console.log("Run the registerUser.js application before retrying");
        //   return;
        //   // return {
        //   //   msg: "User registration failed.",
        //   // };
        // }
        // Create a new gateway for connecting to our peer node.
        // const wallet = await new FileSystemWallet(walletPath);
        const gateway = new Gateway();
        await gateway.connect(ccp, {
            wallet,
            identity: userId,
            discovery: {
                enabled: true,
                asLocalhost: false
            },
        });
        // await gateway.connect(ccp, { wallet, identity: userdata.akcessId, discovery: { enabled: false, asLocalhost: true } });

        // Get the network (channel) our contract is deployed to.
        // const network = await gateway.getNetwork('akcesschannel');
        const network = await gateway.getNetwork(channelName);

        // Get the contract from the network.
        // const contract = network.getContract('akcess');
        const contract = network.getContract(contractName);

        // Submit the specified transaction.
        const result = await contract.evaluateTransaction(functionName, channelName);

        const root = await protobuf.load(__dirname + '/blockinfo.proto');
        var BlockchainInfo = root.lookupType('blockinfo.BlockchainInfo');

        var response = BlockchainInfo.decode(result);
        var maxBlockNum = response.height - 1;

        await blocks.countDocuments({}, async function (err, count) {
            updateCount(err, count);
            return
        });

        var currentBlocks;
        var error;

        function updateCount(err, count) {
            if (err) {
                console.log("error while fetching blocks", err);
                error = err;
            }
            currentBlocks = count - 1;
        }
        if (error) {
            return {
                success: false,
                message: "error while fetching blocks"
            }
        }
        var currentHash = response.currentBlockHash.toString('hex');
        for (let i = maxBlockNum; i > currentBlocks; i--) {
            const result = await contract.evaluateTransaction("GetBlockByNumber", channelName, i);
            const blockResponse = BlockDecoder.decode(result);
            const block_num = blockResponse.header.number;
            const txcount = blockResponse.data.data.length;
            const dataHash = blockResponse.header.data_hash;
            const previousHash = blockResponse.header.previous_hash;
            const transactions = [];
            var createdAt;
            for (var j = 0; j < txcount; j++) {
                let tx_id = blockResponse.data.data[j].payload.header.channel_header.tx_id;
                let timestamp = blockResponse.data.data[j].payload.header.channel_header.timestamp;
                if(j == 0){
                    createdAt = timestamp;
                }
                let transaction = {
                    txId: tx_id,
                    timestamp: timestamp,
                }
                transactions.push(transaction);
               
            }
            const blockEntry = await new blocks({
                blockNum: block_num,
                txCount: txcount,
                dataHash: dataHash,
                blockHash: currentHash,
                preHash: previousHash,
                transactions: transactions,
                createdDate: new Date(createdAt),
            });

            await blockEntry
                .save()
                .then((result, error) => {
                    console.log("Block " + block_num + " recorded successfully");
                })
                .catch((error) => {
                    console.log("ERROR DB", error);
                });
            currentHash = previousHash;
        }
        var resp = {
            success: true,
            data: maxBlockNum,
        }
        await gateway.disconnect();
        return resp;
    } catch (exception) {
        // logger.error(exception.errors);
        return exception;
    }
};


exports.FabricGetBlocksTransaction = async (
    channelName,
    contractName,
    functionName,
    userId,
    userKey,
    txId,
) => {
    try {
        // load the network configuration
        const ccpPath = path.resolve(
            __dirname,
            "connection-profile",
            "connection-busy.json"
        );
        const ccp = JSON.parse(fs.readFileSync(ccpPath, "utf8"));
        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(
            process.cwd(),
            "blockchain",
            "network",
            "wallet"
        );
        // const walletPath = path.resolve(__dirname, '..', '..', 'network', 'wallet')
        const wallet = await Wallets.newFileSystemWallet(walletPath);


        await wallet.put(userId, userKey);

        // Check to see if we've already enrolled the user.
        // const identity = await wallet.exists(arguements.akcessId);
        // const identity = await wallet.get(userId);
        // if (!identity) {
        //   console.log("An identity for the user does not exist in the wallet");
        //   console.log("Run the registerUser.js application before retrying");
        //   return;
        //   // return {
        //   //   msg: "User registration failed.",
        //   // };
        // }
        // Create a new gateway for connecting to our peer node.
        // const wallet = await new FileSystemWallet(walletPath);
        const gateway = new Gateway();
        await gateway.connect(ccp, {
            wallet,
            identity: userId,
            discovery: {
                enabled: true,
                asLocalhost: false
            },
        });
        // await gateway.connect(ccp, { wallet, identity: userdata.akcessId, discovery: { enabled: false, asLocalhost: true } });

        // Get the network (channel) our contract is deployed to.
        // const network = await gateway.getNetwork('akcesschannel');
        const network = await gateway.getNetwork(channelName);

        // Get the contract from the network.
        // const contract = network.getContract('akcess');
        const contract = network.getContract(contractName);

        const result = await contract.evaluateTransaction(functionName, channelName, txId);
        const blockResponse = BlockDecoder.decode(result);
        
        const txcount = blockResponse.data.data.length;
        var transactionTimestamp;
        // retrieving the transaction timestamp
        for (var j = 0; j < txcount; j++) {
            let tx_id = blockResponse.data.data[j].payload.header.channel_header.tx_id;
            if(tx_id == txId){
                transactionTimestamp = blockResponse.data.data[j].payload.header.channel_header.timestamp;
            }
        }
        const response = {
            blockNum : blockResponse.header.number,
            dataHash: blockResponse.header.data_hash,
            timestamp: transactionTimestamp,
        }
        await gateway.disconnect();
        return response;
    } catch (exception) {
        // logger.error(exception.errors);
        return exception;
    }
};