const User = require("../../models/Users");
const transactions = require("../../models/transactions");
const submitTransaction = require("../../../blockchain/submitTransaction");
const NftTokens = require("../../models/nft-token");
const constants = require("../../../constants");
const {
  Certificate
} = require("@fidm/x509");
const bs58 = require("bs58");


module.exports = async (req, res, next) => {
  try {
    const walletId = req.body.walletId,
      tokenSymbol = req.body.tokenSymbol,
      metaData = req.body.metaData,
      blockchain_credentials = req.body.credentials;

    const user = await User.findOne({
      walletId: walletId
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
        "BusyTokens:UpdateTokenMetaData",
        user.userId,
        blockchain_credentials,
        tokenSymbol,
        JSON.stringify(metaData)
      );

      console.log(response);
      const resp = JSON.parse(response);
      if (resp.success == true) {
        var query = await NftTokens.findOne({
          tokenSymbol: tokenSymbol,
        });
        const transactionEntry = await new transactions({
          transactionType: "UpdateTokenMetaData",
          transactionId: resp.txId,
          submitTime: new Date(),
          payload: {
            address: walletId,
            token: tokenSymbol,
            newMetaData: metaData,
            tokenAddress: query.tokenAddress,
          },
          status: "submitted"
        });

        await transactionEntry
          .save()
          .then((result, error) => {
            console.log("update token metadata is updated");
          })
          .catch((error) => {
            console.log("ERROR DB", error);
          });   
        return res.send(200, {
          status: true,
          message: "Token metadata updated successfully",
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
      console.log("WalletId do not exists.");
      return res.send(404, {
        status: false,
        message: `WalletId does not exist`,
      });
    }
  } catch (exception) {
    console.log(exception);
    return res.send(500, {
      status: false,
      message: exception.message,
    });
  }
};