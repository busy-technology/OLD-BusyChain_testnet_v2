const Admin = require("../../models/admin");
const User = require("../../models/Users");
const constants = require("../../../constants");
const submitTransaction = require("../../../blockchain/submitTransaction");
const transactions = require("../../models/transactions");

module.exports = async (req, res, next) => {
  try {
    const recipient = req.body.recipient,
      amount = req.body.amount,
      numerator = req.body.numerator,
      denominator = req.body.denominator,
      releaseAt = req.body.releaseAt,
      adminId = "busy_network";
    var userId = "sample";

    console.log("IN USER");
    const adminData = await Admin.findOne({
      userId: adminId
    });
    console.log("ADMIN", adminData);

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
      walletId: recipient
    });
    console.log("User", user);
    if (user) {
      userId = user.userId;
      const response = await submitTransaction.SubmitTransaction(
        constants.BUSY_CHANNEL_NAME,
        constants.DEFAULT_CONTRACT_NAME,
        "MultibeneficiaryVestingV1",
        userId,
        blockchain_credentials,
        recipient,
        amount,
        numerator,
        denominator,
        releaseAt
      );
      const resp = JSON.parse(response);
      const txId = resp.txId;

      if (resp.success == true) {
        const transactionEntry = await new transactions({
          transactionType: "vestingv1",
          transactionId: resp.txId,
          submitTime: new Date(),
          payload: {
            recipient: recipient,
            amount: amount,
            numerator: numerator,
            denominator: denominator,
            releaseAt: releaseAt,
          },
          status: "submitted"
        });

        await transactionEntry
          .save()
          .then((result, error) => {
            console.log("vestingv1 transaction recorded.");
          })
          .catch((error) => {
            console.log("ERROR DB", error);
          });

        return res.send(200, {
          status: true,
          message: "Request to create vesting has been successfully accepted",
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
        message: `Wallet does not exist`,
      });
    }
  } catch (exception) {
    console.log("EXCEPTION", exception);
    return res.send(500, {
      status: false,
      message: exception.message,
    });
  }
};