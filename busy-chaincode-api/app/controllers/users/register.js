const User = require("../../models/Users");
const registerUser = require("../../../blockchain/registerUser");
const bcrypt = require("bcrypt");
const saltRounds = 10;
const submitTransaction = require("../../../blockchain/submitTransaction");
const bs58 = require("bs58");
const bip39 = require("bip39");
const constants = require("../../../constants");


module.exports = async (req, res, next) => {
  try {
    const userId = req.body.userId,
      firstName = req.body.firstName || "",
      lastName = req.body.lastName || "",
      email = req.body.email || "",
      mobile = req.body.mobile || "",
      password = req.body.password,
      country = req.body.country || "",
      confirmPassword = req.body.confirmPassword;

    const user = await User.findOne({
      userId: userId
    });
    console.log("User", user);
    if (user) {
      console.log("UserId already taken.");
      return res.send(409, {
        status: false,
        message: `Nickname is already taken`,
      });
    } else if (password != confirmPassword) {
      console.log("Passwords do not match.");
      return res.send(400, {
        status: false,
        message: "Passwords do not match",
      });
    } else {
      const mnemonic = bip39.generateMnemonic();
      const salt = await bcrypt.genSaltSync(saltRounds);
      const hash = await bcrypt.hashSync(password, salt);
      const registeruser = await registerUser({
        userId: userId
      }, mnemonic);
      if (registeruser) {
        const response = await submitTransaction.SubmitTransaction(
          constants.BUSY_CHANNEL_NAME,
          constants.DEFAULT_CONTRACT_NAME,
          "CreateUser",
          userId,
          registeruser);
        const resp = JSON.parse(response);

        const bytes = Buffer.from(
          registeruser.credentials.privateKey,
          "utf-8"
        );

        const encodedPrivateKey = bs58.encode(bytes);

        if (resp.success == true) {
          const users = await new User({
            firstName: firstName,
            lastName: lastName,
            email: email,
            mobile: mobile,
            userId: userId,
            walletId: resp.data,
            password: hash,
            country: country,
            txId: resp.txId,
            tokens: {
              "BUSY": {
                "balance": 0,
                "createdAt": new Date(),
                "type": "busy"
              },
            },
            messageCoins: {
              totalCoins: 0,
            },
          });

          await users
            .save()
            .then((result, error) => {
              console.log("User registered.");
            })
            .catch((error) => {
              console.log("ERROR DB", error);
            });

          registeruser.credentials.privateKey = encodedPrivateKey;

          return res.send(200, {
            status: true,
            message: "User has been successfully registered",
            seedPhase: mnemonic,
            privateKey: registeruser,
            chaincodeResponse: resp,
          });
        } else {
          console.log("Failed to execute chaincode function");
          return res.send(500, {
            status: false,
            message: `Failed to execute chaincode function`,
          });
        }
      } else {
        console.log("Failed to enroll the user");
        return res.send(500, {
          status: false,
          message: `Failed to enroll the user`,
        });
      }
    }
  } catch (exception) {
    console.log(exception);
    return res.send(500, {
      status: false,
      message: exception.message,
    });
  }
};