const User = require("../../models/Users");
const bs58 = require("bs58");
const constants = require("../../../constants");
const submitTransaction = require("../../../blockchain/submitTransaction");
const transactions = require("../../models/transactions");

const {
  Certificate
} = require("@fidm/x509");


module.exports = async (req, res, next) => {
  const sender = req.body.sender;
  const recipient = req.body.recipient;
  const blockchain_credentials = req.body.credentials;
  const nftName = req.body.nftName;
  try {
    const sendUser = await User.findOne({
      walletId: sender
    });
    const recUser = await User.findOne({
      walletId: recipient
    })
    if (sendUser && recUser) {
      const commanName = Certificate.fromPEM(
        Buffer.from(blockchain_credentials.credentials.certificate, "utf-8")
      ).subject.commonName;
      console.log("CN", commanName);
      if (sendUser.userId != commanName) {
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
        "BusyNFT:Transfer",
        sendUser.userId,
        blockchain_credentials,
        sender,
        recipient,
        nftName,
      );
      const resp = JSON.parse(response);
      if (resp.success == true) {
        console.log("Message Sent Successfully")
        
        const transactionEntry = await new transactions({
          transactionType: "busynftTransfer",
          transactionId: resp.txId,
          submitTime: new Date(),
          payload: {
            sender: sender,
            recipient: recipient,
            nftName: nftName,
          },
          status: "submitted"
        });
       
        await transactionEntry
          .save()
          .then((result, error) => {
            console.log("Token Transferred recorded.");
          })
          .catch((error) => {
            console.log("ERROR DB", error);
          });

        return res.send(200, {
          status: true,
          message: "Request to transfer NFT token has been successfully accepted",
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
      if (!sendUser) {
        console.log("Sender WalletId do not exist.");
        return res.send(404, {
          status: false,
          message: `Sender’s address does not exist`,
        });
      } else {
        console.log("Recipient WalletId do not exist.");
        return res.send(404, {
          status: false,
          message: `Recipient's address does not exist`,
        });
      };
    }
  } catch (exception) {
    console.log(exception);
    return res.send(404, {
      status: false,
      message: exception.message,
    });
  };
};