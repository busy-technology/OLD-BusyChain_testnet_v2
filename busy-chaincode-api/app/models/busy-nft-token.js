const mongoose = require("mongoose");
const Schema = mongoose.Schema;

const busynftTokensSchema = new Schema({
  nftName: {
    type: String,
    required: true,
  },
  tokenAdmin: {
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
module.exports = mongoose.model("busyNftTokens", busynftTokensSchema);
