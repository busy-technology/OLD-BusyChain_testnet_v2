const User = require("../../models/Users");
const Pool = require("../../models/Pools");
const transactions = require("../../models/transactions");
const bs58 = require("bs58");
const constants = require("../../../constants");
const submitTransaction = require("../../../blockchain/submitTransaction");
const {
  Certificate
} = require("@fidm/x509");

module.exports = async (req, res, next) => {
  const poolName = req.body.poolName;
  const poolDescription = req.body.poolDescription;
  const walletId = req.body.walletId;
  const blockchain_credentials = req.body.credentials;
  try {
    const user = await User.findOne({
      walletId: walletId
    });
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

      const decodedPrivateKey = bs58.decode(
        blockchain_credentials.credentials.privateKey
      );

      blockchain_credentials.credentials.privateKey =
        decodedPrivateKey.toString();

      const response = await submitTransaction.SubmitTransaction(
        constants.BUSY_CHANNEL_NAME,
        constants.DEFAULT_CONTRACT_NAME,
        "BusyVoting:CreatePool",
        user.userId,
        blockchain_credentials,
        walletId,
        poolName,
        poolDescription,
        constants.BUSY_TOKEN
      );
      const resp = JSON.parse(response);

      if (resp.success == true) {
        console.log("Pool Created Successfully")

        const poolEntry = await new Pool({
          PoolID: resp.txId,
          PoolInfo: resp.data,
        });

        await poolEntry
          .save()
          .then((result, error) => {
            console.log("Pool info is save in database");
          })
          .catch((error) => {
            console.log("ERROR DB", error);
          });

        const transactionEntry = await new transactions({
          transactionType: "createpool",
          transactionId: resp.txId,
          submitTime: new Date(),
          payload: {
            tokenName: constants.BUSY_TOKEN,
            amount: resp.data.poolFee,
            sender: walletId,
            receiver: resp.txId,

          },
          status: "submitted"
        });

        await transactionEntry
          .save()
          .then((result, error) => {
            console.log("Create Pool transaction recorded.");
          })
          .catch((error) => {
            console.log("ERROR DB", error);
          });

        return res.send(200, {
          status: true,
          message: "Request to create a new voting pool has been successfully accepted",
          chaincodeResponse: resp,
        })
      } else {
        console.log("Failed to execute chaincode function");
        return res.send(500, {
          status: false,
          message: resp.message,
        });
      };
    } else {
      console.log("WalletId do not exists.");
      return res.send(404, {
        statPus: false,
        message: `Wallet does not exist`,
      });
    }
  } catch (exception) {
    console.log(exception);
    return res.send(500, {
      status: false,
      message: exception.message
    });
  };
};