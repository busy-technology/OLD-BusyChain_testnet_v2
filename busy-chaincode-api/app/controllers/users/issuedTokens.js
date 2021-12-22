const IssuetokenTransactions = require("../../models/issued-tokens");
const transactions = require("../../models/transactions");

module.exports = async (req, res, next) => {
  var query = await IssuetokenTransactions.find({});
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
      tokenName: query[i].tokenName,
      tokenAdmin: query[i].tokenAdmin,
      tokenDecimals: query[i].tokenDecimals,
      tokenSymbol: query[i].tokenSymbol,
      tokenAdmin: query[i].tokenAdmin,
      tokenId: query[i].tokenId,
      tokenSupply: query[i].tokenSupply,
      logoPath: query[i].logoPath,
      websiteUrl: query[i].websiteUrl,
      socialMedia: query[i].socialMedia,
      createdDate: query[i].createdDate,
    };
    var transaction = await transactions.find({
      transactionId: query[i].transactionId,
      status: "VALID"
    });

    if (transaction) {
      output.push(object);
    }
  }

  return res.send(200, {
    status: true,
    message: "Issued tokens have been successfully fetched",
    output: output,
  });
}; 
