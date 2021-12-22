const Admin = require("../../models/admin");
const transactions = require("../../models/transactions");
const User = require("../../models/Users");
const constants = require("../../../constants");
const submitTransaction = require("../../../blockchain/submitTransactionWaitBlockCommit");
const blocks = require("../../../blockchain/blocks");

module.exports = async (req, res, next) => {
  try {
    const address = req.body.walletId,
      token = req.body.token,
      amount = req.body.amount,
      adminId = "busy_network",
      userId = "sample";

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
      walletId: address
    });

    if (user) {
      const response = await submitTransaction.SubmitTransaction(
        constants.BUSY_CHANNEL_NAME,
        constants.DEFAULT_CONTRACT_NAME,
        "Burn",
        userId,
        blockchain_credentials,
        address,
        amount,
        token
      );
      const resp = JSON.parse(response);
      console.log("DATA 2", resp);
      const txId = resp.txId;

      if (resp.success == true) {
        const blockinfo = await blocks.FabricGetBlocksTransaction(
          constants.BUSY_CHANNEL_NAME,
          constants.QSCC_CONTRACT_NAME,
          constants.GET_BLOCKS_FUNCTION_NAME,
          userId,
          blockchain_credentials,
          resp.txId
        )

        const transactionEntry = await new transactions({
          transactionType: "burn",
          transactionId: txId,
          submitTime: blockinfo.timestamp,
          payload: {
            amount: amount,
            address: address,
            token: token
          },
          status: "VALID",
          blockNum: blockinfo.blockNum,
          dataHash: blockinfo.dataHash
        });

        await transactionEntry
          .save()
          .then((result, error) => {
            console.log("Burn Token transaction recorded.");
          })
          .catch((error) => {
            console.log("ERROR DB", error);
          });

        return res.send(200, {
          status: true,
          message: "Burn has been successfully completed",
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
      console.log("WalletId does not exists.");
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