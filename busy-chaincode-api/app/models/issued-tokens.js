const mongoose = require("mongoose");
const Schema = mongoose.Schema;

const issuedTokensSchema = new Schema({
  tokenName: {
    type: String,
    required: true,
  },
  tokenSymbol: {
    type: String,
    required: true,
  },
  tokenDecimals: {
    type: Number,
    required: true,
  },
  transactionId: {
    type: String,
    required: true,
  },
  tokenAdmin: {
    type: String,
    required: true,
  },
  tokenId: {
    type: String,
    required: true,
  },
  tokenSupply: {
    type: String,
    required: true,
  },
  logoPath: {
    type: String,
  },
  websiteUrl: {
    type: String,
  },
  socialMedia: {
    type: String,
  },
  createdDate: {
    type: Date,
    required: true,
  }
});

module.exports = mongoose.model("issuedTokens", issuedTokensSchema);
