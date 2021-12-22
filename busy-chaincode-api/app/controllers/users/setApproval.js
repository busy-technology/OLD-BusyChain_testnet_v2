const User = require("../../models/Users");
const bs58 = require("bs58");
const constants = require("../../../constants");
const submitTransaction = require("../../../blockchain/submitTransaction");
const transactions = require("../../models/transactions");

const {
  Certificate
} = require("@fidm/x509");


module.exports = async (req, res, next) => {
  const walletId = req.body.walletId;
  const blockchain_credentials = req.body.credentials;
  const operator = req.body.operator;
  const approved = req.body.approved;
  try {
    const user = await User.findOne({
      walletId: walletId
    });
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
        "BusyTokens:SetApprovalForAll",
        user.userId,
        blockchain_credentials,
        operator,
        approved
      );
      const resp = JSON.parse(response);
      if (resp.success == true) {
        console.log("Nft Approval Succesffuly Set")
        
        const transactionEntry = await new transactions({
          transactionType: "nftApprovalSet",
          transactionId: resp.txId,
          submitTime: new Date(),
          payload: {
            sender: walletId,
            operator: operator,
            approved: approved,
          },
          status: "submitted"
        });
       
        await transactionEntry
          .save()
          .then((result, error) => {
            console.log("Set NFT Approval transaction recorded.");
          })
          .catch((error) => {
            console.log("ERROR DB", error);
          });

        return res.send(200, {
          status: true,
          message: "Request to set NFT approval has been successfully accepted",
          chaincodeResponse: resp,
        })
      } else {
        console.log("Failed to execute chaincode function");
        return res.send(500, {
          status: false,
          message: resp.message,
        });
      };
    } else {
        console.log("WalletId do not exist.");
        return res.send(404, {
          status: false,
          message: `WalletId does not exist`,
        });
    }
  } catch (exception) {
    console.log(exception);
    return res.send(404, {
      status: false,
      message: exception.message,
    });
  };
};