const {
  Gateway,
  Wallets,
  FileSystemWallet
} = require("fabric-network");
const fs = require("fs");
const path = require("path");


exports.QueryTransaction = async (
  channelName,
  contractName,
  functionName,
  ...args
) => {
  try {

    console.log("Recieved a Query Transaction for ", functionName);
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

    // fetching the wallets
    const identity = await wallet.get(args[0]);

    // Put the creds in the filesystem if it does not exists
    if (!identity) {
      await wallet.put(args[0], args[1]);
    }

    // Create a new gateway for connecting to our peer node.
    const gateway = new Gateway();
    await gateway.connect(ccp, {
      wallet,
      identity: args[0],
      discovery: {
        enabled: true,
        asLocalhost: false
      },
    });

    // Retreiving the required args from the function
    var evaluateArgs = [];
    for (let i = 2; i < args.length; i++) {
      evaluateArgs.push(args[i]);
    }

    console.log("Evaluting the transaction on channel", channelName)
    // Get the network (channel) our contract is deployed to.
    // const network = await gateway.getNetwork('akcesschannel');
    const network = await gateway.getNetwork(channelName);

    // Get the contract from the network.
    // const contract = network.getContract('akcess');
    const contract = network.getContract(contractName);

    // Submit the specified transaction.
    // const invoked = await contract.submitTransaction('UpdateMobileNo', userdata.akcessId, userdata.phoneNumber);
    const result = await contract.evaluateTransaction(
      functionName,
      ...evaluateArgs
    );

    console.log("Transaction has been submitted");

    // Disconnect from the gateway.
    gateway.disconnect();
    return result.toString();
  } catch (exception) {
    throw exception;
  }
};