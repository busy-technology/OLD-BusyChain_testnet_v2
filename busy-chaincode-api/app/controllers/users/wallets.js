const Wallet = require("../../models/Wallets");

module.exports = async (req, res, next) => {
  var query = await Wallet.find({
    amount: {
      $ne: "0"
    }
  });
  if (!query) {
    return res.send(500, {
      status: false,
      message: "error fetching the transations",
    });
  }
  const output = [];

  for (let i = 0; i < query.length; i++) {
    var object = {
      walletId: query[i].stakingWalletId,
      createdDate: query[i].createdDate,
      createdFrom: query[i].walletId,
    };
    output.push(object);
  }

  return res.send(200, {
    status: true,
    message: "Staking addresses have been successfully fetched",
    output: output,
  });
};