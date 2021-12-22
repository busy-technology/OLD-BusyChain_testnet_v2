const User = require("../../models/Users");
const Wallet = require("../../models/Wallets");
const {
  Certificate
} = require("@fidm/x509");
const bs58 = require("bs58");
const constants = require("../../../constants");
const submitTransaction = require("../../../blockchain/submitTransaction");
const transactions = require("../../models/transactions");

module.exports = async (req, res, next) => {
  try {
    const stakingAddr = req.body.stakingAddr,
      blockchain_credentials = req.body.credentials;

    const address = await Wallet.findOne({
      stakingWalletId: stakingAddr,
    });
    console.log("ADDRESS", address);

    if (address) {
      const user = await User.findOne({
        userId: address.userId
      });
      console.log("User", user);

      if (user) {
        const commanName = Certificate.fromPEM(
          Buffer.from(blockchain_credentials.credentials.certificate, "utf-8")
        ).subject.commonName;
        console.log("CN", commanName);

        if (user.userId != commanName) {
          return res.send(404, {
            status: false,
            message: `User’s certificate is not valid`,
          });
        }
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

        const decodedPrivateKey = bs58.decode(
          blockchain_credentials.credentials.privateKey
        );

        blockchain_credentials.credentials.privateKey =
          decodedPrivateKey.toString();

        const response = await submitTransaction.SubmitTransaction(
          constants.BUSY_CHANNEL_NAME,
          constants.DEFAULT_CONTRACT_NAME,
          "Unstake",
          user.userId,
          blockchain_credentials,
          stakingAddr
        );
        const resp = JSON.parse(response);
        const txId = response.txId;
        console.log("TRANSACTION ID", txId);

        if (resp.success == true) {
          const transactionEntry = await new transactions({
            transactionType: "unstake",
            transactionId: resp.txId,
            submitTime: new Date(),
            payload: {
              token: constants.BUSY_TOKEN,
              amount: resp.data.amount,
              totalReward: resp.data.totalReward,
              claimed: resp.data.claimed,
              walletId: resp.data.defaultWalletAddress,
              stakingWalletId: resp.data.stakingAddr,
            },
            status: "submitted"
          });
         
          await transactionEntry
            .save()
            .then((result, error) => {
              console.log("Unstake transaction recorded.");
            })
            .catch((error) => {
              console.log("ERROR DB", error);
            });
  
          await Wallet.updateOne({
            stakingWalletId: address.stakingWalletId
          }, {
            amount: "0",
            totalReward: resp.data.totalReward,
            claimed: resp.data.claimed
          });

          return res.send(200, {
            status: true,
            message: "Request to unstake staking address has been successfully accepted",
            chaincodeResponse: resp,
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
        console.log("User having this stakking address not found.");
        return res.send(404, {
          status: false,
          message: `User does not own this staking address`,
        });
      }
    } else {
      console.log("stakingAddr do not exists.");
      return res.send(404, {
        status: false,
        message: `Staking address does not exist`,
      });
    }
  } catch (exception) {
    console.log(exception);
    return res.send(404, {
      status: false,
      message: exception.message,
    });
  }
};