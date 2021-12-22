const {
    Gateway,
    Wallets,
    FileSystemWallet,
    DefaultEventHandlerStrategies
} = require("fabric-network");
const fs = require("fs");
const path = require("path");

exports.SubmitTransaction = async (
    channelName,
    contractName,
    functionName,
    ...args
) => {
    try {
        console.log("Recieved a Submit Transaction for ", functionName);

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

        const identity = await wallet.get(args[0]);

        if (!identity) {
            await wallet.put(args[0], args[1]);
        }

        // Create a new gateway for connecting to our peer node.
        const gateway = new Gateway();
        await gateway.connect(ccp, {
            wallet,
            identity: args[0],
            eventHandlerOptions: {
                strategy: DefaultEventHandlerStrategies.NONE,
            },
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

        console.log("Submitting the transaction on channel", channelName)

        // Retreiving the required args from the function
        var submitArgs = [];
        for (let i = 2; i < args.length; i++) {
            submitArgs.push(args[i]);
        }

        console.log("submit args", submitArgs);
        // Submit the specified transaction.
        const result = await contract.submitTransaction(
            functionName,
            ...submitArgs
        );
        console.log("Transaction has been submitted");

        // Disconnect from the gateway.
        gateway.disconnect();
        return result.toString();
    } catch (exception) {
        if(exception.responses && exception.responses.length != 0){
          throw exception.responses[0].response;
        } else {
            throw exception.message
        }
    }
};