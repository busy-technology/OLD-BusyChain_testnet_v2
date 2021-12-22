const Admin = require("../../models/admin");
const queryTransaction = require("../../../blockchain/queryTransaction");
const constants = require("../../../constants");

module.exports = async (req, res, next) => {
  try {
    //const userId = req.body.userId;
    const userId = "sample";
    const walletId = req.body.walletId;
    const token = req.body.token;
    const adminId = "busy_network";

    //const user = await User.findOne({ userId: userId });

    // if (user) {
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

    // const wallet = await Wallet.findOne({ userId: userId });
    // console.log("WALLET", wallet);

    // if (wallet == null) {
    //   return res.send(404, {
    //     status: false,
    //     message: `Wallet do not exist.`,
    //   });
    // }

    const response = await queryTransaction.QueryTransaction(
      constants.BUSY_CHANNEL_NAME,
      constants.DEFAULT_CONTRACT_NAME,
      "GetBalance",
      userId,
      blockchain_credentials,
      walletId,
      token
    );
    const resp = JSON.parse(response);

    if (resp.success == true) {
      return res.send(200, {
        status: true,
        message: "Balance has been successfully fetched",
        chaincodeResponse: resp,
      });
    } else {
      console.log("Failed to execute chaincode function");
      return res.send(404, {
        status: false,
        message: `Failed to execute chaincode function`,
        chaincodeResponse: resp,
      });
    }
    // } else {
    //   console.log("UserId do not exists.");
    //   return res.send(404, {
    //     status: false,
    //     message: `UserId do not exists.`,
    //   });
    // }
  } catch (exception) {
    console.log(exception);
    return res.send(404, {
      status: false,
      message: exception.message,
    });
  }
};