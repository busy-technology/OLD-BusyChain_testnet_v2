const User = require("../../models/Users");
const Admin = require("../../models/admin");
const submitTransaction = require("../../../blockchain/submitTransaction");
const constants = require("../../../constants");
const transactions = require("../../models/transactions");

module.exports = async (req, res, next) => {
  const adminId = "busy_network";
  const adminData = await Admin.findOne({ userId: adminId });

  const credentials = {
    certificate: adminData.certificate.credentials.certificate,
    privateKey: adminData.certificate.credentials.privateKey,
  };

  const blockchain_credentials = {
    credentials: credentials,
    mspId: adminData.certificate.mspId,
    type: adminData.certificate.type,
  };

  try {
    const user = await User.findOne({ userId: adminId });
    const response = await submitTransaction.SubmitTransaction(
      constants.BUSY_CHANNEL_NAME,
      constants.DEFAULT_CONTRACT_NAME,
      "BusyVoting:DestroyPool",
      adminId, 
      blockchain_credentials
    );
    const resp = JSON.parse(response);
    if (resp.success == true) {
      const transactionEntry = await new transactions({
        transactionType: "destroypool",
        transactionId: resp.txId,
        submitTime: new Date(),
        payload: {
          sender: adminId,  
        },
        status: "submitted"
      });

      await transactionEntry
        .save()
        .then((result, error) => {
          console.log("Destroy Pool transaction recorded.");
        })
        .catch((error) => {
          console.log("ERROR DB", error);
        });

      console.log("Pool Destroyed Successfully");
      return res.send(200, {
        status: true,
        message: "Request to destroy old voting pool has been successfully accepted",
        chaincodeResponse: resp,
      });
    } else {
      console.log("Failed to execute chaincode function");
      return res.send(500, {
        status: false,
        message: resp.message,
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
