const Admin = require("../../models/admin");
const IssuetokenTransactions = require("../../models/issued-tokens");
const queryTransaction = require("../../../blockchain/queryTransaction");
const constants = require("../../../constants");

module.exports = async (req, res, next) => {
  try {
    const symbol = req.body.symbol;
    const adminId = "busy_network";
    const userId = "sample";

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

    // const lowerToken = symbol.toLowerCase();
    // console.log("LOWER TOKEN", lowerToken);

    const coinSymbol = await IssuetokenTransactions.findOne({
      tokenSymbol: symbol,
    });
    console.log("COIN", coinSymbol);

    if (coinSymbol || symbol == "BUSY") {
      const response = await queryTransaction.QueryTransaction(
        constants.BUSY_CHANNEL_NAME,
        constants.DEFAULT_CONTRACT_NAME,
        "GetTotalSupply",
        userId,
        blockchain_credentials,
        symbol
      );
      const resp = JSON.parse(response);
      console.log("BALANCE", resp.data);

      if (resp.success == true) {
        return res.send(200, {
          status: true,
          message: "Total supply has been successfully fetched",
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
      console.log("symbol do not exists.");
      return res.send(404, {
        status: false,
        message: `Symbol does not exist`,
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
