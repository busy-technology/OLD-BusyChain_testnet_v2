const User = require("../../models/Users");
const transactions = require("../../models/transactions");
const submitTransaction = require("../../../blockchain/submitTransaction");
const constants = require("../../../constants");
const busyNftTokens = require("../../models/busy-nft-token");
const {
  Certificate
} = require("@fidm/x509");
const bs58 = require("bs58");


module.exports = async (req, res, next) => {
  try {
    const walletId = req.body.walletId,
      nftName = req.body.nftName,
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
        "BusyNFT:Mint",
        user.userId,
        blockchain_credentials,
        walletId,
        nftName,
        JSON.stringify(metaData),
      );

      console.log(response);
      const resp = JSON.parse(response);
      if (resp.success == true) {

        const tokenEntry = await new busyNftTokens({
          nftName: nftName,
          transactionId: resp.txId,
          tokenAdmin: walletId,
          properties: metaData,
          createdDate: new Date(),
        });

        await tokenEntry
          .save()
          .then((result, error) => {
            console.log("Nft Mint Token saved.");
          })
          .catch((error) => {
            console.log("ERROR DB", error);
          });


        const transactionEntry = await new transactions({
          transactionType: "busymintToken",
          transactionId: resp.txId,
          submitTime: new Date(),
          payload: {
            address: walletId,
            nftName: nftName,
            tokenMetaData: metaData,
          },
          status: "submitted"
        });

        await transactionEntry
          .save()
          .then((result, error) => {
            console.log("mint transaction recorded.");
          })
          .catch((error) => {
            console.log("ERROR DB", error);
          });   
        return res.send(200, {
          status: true,
          message: "Request to mint the new Busy NFT token has been successfully accepted",
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