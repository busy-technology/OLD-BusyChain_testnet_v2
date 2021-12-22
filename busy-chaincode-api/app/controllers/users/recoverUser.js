const User = require("../../models/Users");
const recoverUser = require("../../../blockchain/recoverUser");
const bs58 = require("bs58");

module.exports = async (req, res, next) => {
  try {
    const userId = req.body.userId;
    const mnemonic = req.body.mnemonic;

    console.log("USERID", userId);
    console.log("SEED", mnemonic);
    const user = await User.findOne({
      userId: userId
    });

    if (user) {
      try {
        const response = await recoverUser.FabricUserRecover(userId, mnemonic);

        if (response.blockchain_credentials.credentials) {

          const bytes = Buffer.from(
            response.blockchain_credentials.credentials.privateKey,
            "utf-8"
          );

          const encodedPrivateKey = bs58.encode(bytes);

          response.blockchain_credentials.credentials.privateKey =
            encodedPrivateKey;

          return res.send(200, {
            status: true,
            message: "Successfully recovered User credentials",
            privateKey: response.blockchain_credentials,
          });
        }
      } catch (exception) {
        console.log("exception in User exists", exception);
        return res.send(400, {
          status: false,
          message: `The entered seed phrase is not correct`,
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