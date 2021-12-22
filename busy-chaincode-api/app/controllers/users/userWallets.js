const User = require("../../models/Users");
const Admin = require("../..//models/admin");
const queryTransaction = require("../../../blockchain/queryTransaction");
const constants = require("../../../constants");

module.exports = async (req, res, next) => {
  try {
    const userId = req.body.userId;
    const user = await User.findOne({ userId: userId });
    
    if (user) {
      return res.send(200, {
        status: true,
        message: "User wallet has been successfully fetched",
        chaincodeResponse: {
          tokens: user.tokens,
          messageCoins: user.messageCoins,
        },
      });
    } else {
      console.log("user Does not exists");
      return res.send(500, {
        status: false,
        message: `user Does not exists`,
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
