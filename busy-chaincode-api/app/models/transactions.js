const mongoose = require("mongoose");
const Schema = mongoose.Schema;

const TransactionSchema = new Schema({
  transactionType: {
    type: String,
    required: true,
  },
  status: {
    type: String,
    required: true,
  },
  transactionId: {
    type: String,
    required: true,
  },
  blockNum: {
    type: Number,
  },
  dataHash: {
    type: String,
  },
  payload: {
    type: Object,
    required: true,
  },
  submitTime: {
    type: Date,
    required: true,
  },
  updateTime: {
    type: Date,
  },
});

module.exports = mongoose.model("transactions", TransactionSchema);