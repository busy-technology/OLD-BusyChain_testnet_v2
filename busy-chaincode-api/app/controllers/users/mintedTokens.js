const NftTokens = require("../../models/nft-token");
const transactions = require("../../models/transactions");

module.exports = async (req, res, next) => {
  var query = await NftTokens.find({});
  if (!query) {
    return res.send(500, {
      status: false,
      message: "Error fetching issued Tokens",
    });
  }
  console.log("Number of issued Coins:", query.length);
  console.log("OUTPUT", query);

  const output = [];

  for (let i = 0; i < query.length; i++) {
    var object = {
      tokenSymbol: query[i].tokenSymbol,
      tokenAdmin: query[i].tokenAdmin,
      totalSupply: query[i].totalSupply,
      properties: query[i].properties,
      createdDate: query[i].createdDate,
      tokenAddress: query[i].tokenAddress,
    };

    var transaction = await transactions.find({
      transactionId: query[i].transactionId,
      status: "VALID"
    });

    if (transaction && transaction.length > 0) {
      output.push(object);
    }
  }

  return res.send(200, {
    status: true,
    message: "Minted tokens have been successfully fetched",
    output: output,
  });
};