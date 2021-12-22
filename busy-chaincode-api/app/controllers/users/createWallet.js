const User = require("../../models/Users");
const Wallet = require("../../models/Wallets");
const { Certificate } = require("@fidm/x509");
const bs58 = require("bs58");
const constants = require("../../../constants");
const submitTransaction = require("../../../blockchain/submitTransactionWaitBlockCommit");
const transactions = require("../../models/transactions");
const blocks = require("../../../blockchain/blocks");

module.exports = async (req, res, next) => {
  try {
    const userId = req.body.userId,
      blockchain_credentials = req.body.credentials,
      type = req.body.type;
    console.log("TYPE", type);

    const user = await User.findOne({
      userId: userId,
    });
    console.log("User", user);
    if (user) {
      const commanName = Certificate.fromPEM(
        Buffer.from(blockchain_credentials.credentials.certificate, "utf-8")
      ).subject.commonName;
      console.log("CN", commanName);
      if (userId != commanName) {
        return res.send(404, {
          status: false,
          message: `User’s certificate is not valid`,
        });
      }

      if (type == "online" || type == "offline" || type == "staking") {
        if (
          blockchain_credentials.type != "X.509" ||
          blockchain_credentials.mspId != "BusyMSP"
        ) {
          console.log("type of certificate incorrect.");
          return res.send(400, {
            status: false,
            message: `Incorrect type or MSPID`,
          });
        }

        // const decodedPrivateKey = base58.decode(
        //   blockchain_credentials.credentials.privateKey
        // );
        const decodedPrivateKey = bs58.decode(
          blockchain_credentials.credentials.privateKey
        );

        blockchain_credentials.credentials.privateKey =
          decodedPrivateKey.toString();

        const response = await submitTransaction.SubmitTransaction(
          constants.BUSY_CHANNEL_NAME,
          constants.DEFAULT_CONTRACT_NAME,
          "CreateStakingAddress",
          userId,
          blockchain_credentials
        );
        const resp = JSON.parse(response);
        const txId = resp.txId;

        if (resp.success == true) {
          const stakingWalletId = resp.data.stakingAddr;
          const wallet = await new Wallet({
            userId: userId,
            stakingWalletId: stakingWalletId,
            walletId: user.walletId,
            type: type,
            txId: resp.txId,
            amount: resp.data.amount,
            initialStakingLimit: resp.data.amount,
            totalReward: resp.data.totalReward,
            createdDate: new Date(),
            claimed: resp.data.claimed,
          });

          await wallet
            .save()
            .then((result, error) => {
              console.log("Wallet saved.");
            })
            .catch((error) => {
              console.log("ERROR DB", error);
            });

            const blockinfo = await blocks.FabricGetBlocksTransaction(
              constants.BUSY_CHANNEL_NAME,
              constants.QSCC_CONTRACT_NAME,
              constants.GET_BLOCKS_FUNCTION_NAME,
              userId,
              blockchain_credentials,
              resp.txId
            )

            const transactionEntry = await new transactions({
              transactionType: "createstakingaddress",
              transactionId: resp.txId,
              submitTime: blockinfo.timestamp,
              payload: {
                tokenName: constants.BUSY_TOKEN,
                sender: user.walletId,
                stakingWalletId: stakingWalletId,
                type: type,
                amount: resp.data.amount,
                initialStakingLimit: resp.data.amount,
                totalReward: resp.data.totalReward,
                claimed: resp.data.claimed,
              },
              status: "VALID",
              blockNum: blockinfo.blockNum,
              dataHash: blockinfo.dataHash 
            });
  
            await transactionEntry
              .save()
              .then((result, error) => {
                console.log("Create Staking transaction recorded.");
              })
              .catch((error) => {
                console.log("ERROR DB", error);
              });
  
          return res.send(200, {
            status: true,
            message:
              "Staking address has been successfully created",
            chaincodeResponse: stakingWalletId,
          });
        } else {
          console.log("Failed to execute chaincode function");
          return res.send(500, {
            status: false,
            message: `Failed to execute chaincode function`,
            chaincodeResponse: resp,
          });
        }
      } else {
        console.log("Incorrect type of wallet.");
        return res.send(400, {
          status: false,
          message: `Incorrect type of wallet`,
        });
      }
    } else {
      console.log("UserId do not exists.");
      return res.send(404, {
        status: false,
        message: `User does not exist`,
      });
    }
  } catch (exception) {
    var errorMessage = exception.message;
    if(errorMessage == null || errorMessage == ''  || errorMessage ==undefined){
      errorMessage = "staking transaction failed, please try again."
    }
    return res.send(500, {
      status: false,
      message: errorMessage,
    });

  }
};
