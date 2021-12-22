const User = require("../../models/Users");
const {
  Certificate
} = require("@fidm/x509");
const transactions = require("../../models/transactions");
const bs58 = require("bs58");
const constants = require("../../../constants");
const submitTransaction = require("../../../blockchain/submitTransaction");

module.exports = async (req, res, next) => {
  try {
    const sender = req.body.sender,
      blockchain_credentials = req.body.credentials,
      recipiant = req.body.recipiant,
      amount = req.body.amount,
      token = req.body.token;

    const user = await User.findOne({
      walletId: sender,
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

      const receiver = await User.findOne({
        walletId: recipiant,
      });

      if (receiver) {
        const decodedPrivateKey = bs58.decode(
          blockchain_credentials.credentials.privateKey
        );

        blockchain_credentials.credentials.privateKey =
          decodedPrivateKey.toString();


        if (sender == receiver){
          return res.send(409, {
            status: false,
            message: `Sender cannot be same as recipient`,
          });
        }

        const response = await submitTransaction.SubmitTransaction(
          constants.BUSY_CHANNEL_NAME,
          constants.DEFAULT_CONTRACT_NAME,
          "Transfer",
          sender,
          blockchain_credentials,
          recipiant,
          amount,
          token
        );
        const resp = JSON.parse(response);
        const txId = resp.txId;
        console.log("TRANSACTION ID", txId);
       
        if (resp.success == true) {
          
          const transactionEntry = await new transactions({
            transactionType: "transfer",
            transactionId: resp.txId,
            submitTime: new Date(),
            payload: {
               token: token,
               sender: sender,
               recipiant: recipiant,
               amount: amount
            },
            status: "submitted"
          });
         
          await transactionEntry
            .save()
            .then((result, error) => {
              console.log("Transfer transaction recorded.");
            })
            .catch((error) => {
              console.log("ERROR DB", error);
            });
  
          return res.send(200, {
            status: true,
            message: "Request to transfer has been successfully accepted",
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
        console.log("recipient walletId do not exists.");
        return res.send(404, {
          status: false,
          message: `Recipient's address does not exist`,
        });
      }
    } else {
      console.log("sender walletId do not exists.");
      return res.send(404, {
        status: false,
        message: `Sender’s address does not exist`,
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