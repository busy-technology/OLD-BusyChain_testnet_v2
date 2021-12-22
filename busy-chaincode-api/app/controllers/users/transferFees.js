const Admin = require("../../models/admin");
const submitTransaction = require("../../../blockchain/submitTransaction");
const constants = require("../../../constants");
const transactions = require("../../models/transactions");

module.exports = async (req, res, next) => {
  try {
    const newTransferFee = req.body.newTransferFee;
    const adminId = "busy_network";
    const userId = "sample";

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

    const response = await submitTransaction.SubmitTransaction(
      constants.BUSY_CHANNEL_NAME,
      constants.DEFAULT_CONTRACT_NAME,
      "UpdateTransferFee",
      userId,
      blockchain_credentials,
      newTransferFee
    );
    const resp = JSON.parse(response);

    if (resp.success == true) {
      const transactionEntry = await new transactions({
        transactionType: "transferfee",
        transactionId: resp.txId,
        submitTime: new Date(),
        payload: {
          transferfee: newTransferFee,
        },
        status: "submitted"
      });

      await transactionEntry
        .save()
        .then((result, error) => {
          console.log("Transfer Fees transaction recorded.");
        })
        .catch((error) => {
          console.log("ERROR DB", error);
        });

      return res.send(200, {
        status: true,
        message: "Request to update transaction fee has been successfully accepted",
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
  } catch (exception) {
    console.log(exception);
    return res.send(500, {
      status: false,
      message: exception.message,
    });
  }
};