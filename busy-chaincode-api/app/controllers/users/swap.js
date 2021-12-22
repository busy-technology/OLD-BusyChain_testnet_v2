const User = require("../../models/Users");
const Admin = require("../../models/admin");
const transactions = require("../../models/transactions");
const submitTransaction = require("../../../blockchain/submitTransaction");
const constants = require("../../../constants");

module.exports = async (req, res, next) => {
  try {
    const recipiant = req.body.recipiant,
      amount = req.body.amount,
      token = req.body.token,
      adminId = "busy_network";

    const adminData = await Admin.findOne({
      userId: adminId
    });

    const credentials = {
      certificate: adminData.certificate.credentials.certificate,
      privateKey: adminData.certificate.credentials.privateKey,
    };

    const blockchain_credentials = {
      credentials: credentials,
      mspId: adminData.certificate.mspId,
      type: adminData.certificate.type,
    };

    const user = await User.findOne({
      walletId: recipiant
    });
    console.log("User", user);
    if (user) {
      const response = await submitTransaction.SubmitTransaction(
        constants.BUSY_CHANNEL_NAME,
        constants.DEFAULT_CONTRACT_NAME,
        "Transfer",
        adminId,
        blockchain_credentials,
        recipiant,
        amount,
        token
      );

      console.log(response);
      const resp = JSON.parse(response);
      if (resp.success == true) {
        const transactionEntry = await new transactions({
          transactionType: "swap",
          transactionId: resp.txId,
          submitTime: new Date(),
          payload: {
            amount: amount,
            address: recipiant,
            token: token
          },
          status: "submitted"
        });

        await transactionEntry
          .save()
          .then((result, error) => {
            console.log("Swap transaction recorded.");
          })
          .catch((error) => {
            console.log("ERROR DB", error);
          });   
        return res.send(200, {
          status: true,
          message: "Request to Swap has been successfully accepted",
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
      console.log("Recipient WalletId do not exists.");
      return res.send(404, {
        status: false,
        message: `Recipient's address does not exist`,
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