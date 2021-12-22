const Admin = require("../..//models/admin");
const User = require("../../models/Users");
const queryTransaction = require("../../../blockchain/queryTransaction");
const constants = require("../../../constants");

module.exports = async (req, res, next) => {
  try {
    const userIdentity = req.body.userId;
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

    const id = await User.findOne({
      userId: userIdentity,
    });
    //console.log("ADDRESS", address);

    if (id) {
      const response = await queryTransaction.QueryTransaction(
        constants.BUSY_CHANNEL_NAME,
        constants.DEFAULT_CONTRACT_NAME,
        "GetStakingInfo",
        userId,
        blockchain_credentials,
        userIdentity
      );
      const resp = JSON.parse(response);

      if (resp.success == true) {
        return res.send(200, {
          status: true,
          message: "Staking information has been successfully fetched",
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
      console.log("userId do not exists.");
      return res.send(404, {
        status: false,
        message: `User does not exist`,
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