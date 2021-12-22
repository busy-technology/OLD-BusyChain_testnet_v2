const User = require("../../models/Users");
const Wallet = require("../../models/Wallets");
const bs58 = require("bs58");
const queryTransaction = require("../../../blockchain/queryTransaction");
const constants = require("../../../constants");

module.exports = async (req, res, next) => {
  const userId = req.body.userId,
    blockchain_credentials = req.body.credentials;

  const user = await User.findOne({
    userId: userId
  });
  console.log("User", user);
  if (user) {
    const wallet = await Wallet.findOne({
      userId: userId
    });
    if (wallet) {
      const commanName = Certificate.fromPEM(
        Buffer.from(blockchain_credentials.credentials.certificate, "utf-8")
      ).subject.commonName;
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

      const response = await queryTransaction.QueryTransaction(
        constants.BUSY_CHANNEL_NAME,
        constants.DEFAULT_CONTRACT_NAME,
        "GetBalance",
        userId,
        blockchain_credentials,
        wallet.walletId
      );

      const resp = JSON.parse(response);

      if (resp.success == true) {
        return res.send(200, {
          status: true,
          message: "Balance has been successfully fetched",
          chaincodeResponse: resp.chaincodeResponse,
        });
      } else {
        console.log("Failed to execute chaincode function");
        return res.send(500, {
          status: false,
          message: `Failed to execute chaincode function`,
        });
      }
    } else {
      console.log("Wallet do not exists.");
      return res.send(404, {
        status: false,
        message: `Wallet does not exist`,
      });
    }
  } else {
    console.log("UserId do not exists.");
    return res.send(404, {
      status: false,
      message: `User does not exist`,
    });
  }
};