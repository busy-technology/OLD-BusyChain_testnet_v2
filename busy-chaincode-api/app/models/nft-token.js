const mongoose = require("mongoose");
const Schema = mongoose.Schema;

const nftTokensSchema = new Schema({
  tokenSymbol: {
    type: String,
    required: true,
  },
  tokenAdmin: {
    type: String,
    required: true,
  },
  totalSupply: {
      type: String,
      required: true,
  },
  tokenAddress: {
    type: String,
    required: true,
  },
  properties: {
     type: Object,
     required: true,
  },
  transactionId: {
    type: String,
    required: true,
  },
  createdDate: {
    type: Date,
    required: true,
  }
});
module.exports = mongoose.model("nftTokens", nftTokensSchema);
