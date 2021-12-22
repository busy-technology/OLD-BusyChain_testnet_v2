const Admin = require("../../models/admin");
const User = require("../../models/Users");
const queryTransaction = require("../../../blockchain/queryTransaction");
const constants = require("../../../constants");

module.exports = async (req, res, next) => {
  try {
    const address = req.body.walletId;
    // const addressString = address.toString();
    // console.log("address", addressString);
    const adminId = "busy_network";
    var userId = "sample";

    console.log("IN USER");
    const adminData = await Admin.findOne({ userId: adminId });
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

    const user = await User.findOne({ walletId: address });
    console.log("User", user);
    if (user) {
      userId = user.userId;
      const response = await queryTransaction.QueryTransaction(
        constants.BUSY_CHANNEL_NAME,
        constants.DEFAULT_CONTRACT_NAME,
        "GetLockedTokens",
        userId,
        blockchain_credentials,
        address
      );
      const resp = JSON.parse(response);
    
      if (resp.success == true) {
        return res.send(200, {
          status: true,
          message: "Locked tokens fetched.",
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
    console.log(exception);
    return res.send(500, {
      status: false,
      message: exception.message,
    });
  }
};
