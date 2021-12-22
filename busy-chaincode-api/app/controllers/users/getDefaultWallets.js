const Users = require("../../models/Users");
module.exports = async (req, res, next) => {
  var count = await Users.countDocuments({});
  var query = await Users.find({});
  if (!query) {
    return res.send(500, {
      status: false,
      message: err,
    });
  }
  console.log("Number of wallets:", count);
  const output = [];

  for (let i = 0; i < count; i++) {
    var object = {
      walletId: query[i].walletId,
      createdDate: query[i].createdAt,
      tokens: query[i].tokens,
      messageCoins: query[i].messageCoins,
    };
    output.push(object);
  }

  return res.send(200, {
    status: true,
    message: "Default wallets have been successfully fetched",
    output: output,
  });
};