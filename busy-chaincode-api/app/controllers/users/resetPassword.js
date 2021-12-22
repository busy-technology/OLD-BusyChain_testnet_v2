const User = require("../../models/Users");
const bcrypt = require("bcrypt");
const saltRounds = 10;
const queryTransaction = require("../../../blockchain/queryTransaction");
const {
  Certificate
} = require("@fidm/x509");
const bs58 = require("bs58");
const constants = require("../../../constants");


module.exports = async (req, res, next) => {
  try {
    const userId = req.body.userId;
    const newPassword = req.body.newPassword;
    const blockchain_credentials = req.body.credentials;

    console.log("Reset Password for USERID", userId);
    const user = await User.findOne({
      userId: userId,
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

        const response = await queryTransaction.QueryTransaction(
          constants.BUSY_CHANNEL_NAME,
          constants.DEFAULT_CONTRACT_NAME,
          "AuthenticateUser",
          userId,
          blockchain_credentials,
          userId,
        );
        const resp = JSON.parse(response);
        console.log(resp)
        if (resp.success == true) {
          const salt = await bcrypt.genSaltSync(saltRounds);
          const hash = await bcrypt.hashSync(newPassword, salt);
          console.log("NEW HASHED PASSWORD", hash);

          const doc = await User.findOneAndUpdate({
            userId: userId,
          }, {
            password: hash,
          }, {
            upsert: true,
            useFindAndModify: false,
          });
          return res.send(200, {
            status: true,
            message: `Password has been updated successfully`,
          });
        } else {
          console.log("Failed to execute chaincode function");
          return res.send(400, {
            status: false,
            message: resp.message,
          });
        }
    } else {
      console.log("UserId do not exists.");
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