const mongoose = require("mongoose");
const timeStamp = require("mongoose-timestamp");

const WalletSchema = new mongoose.Schema({
  userId: {
    type: String,
    required: true,
    trim: true,
  },
  walletId: {
    type: String,
    required: true,
    trim: true,
  },
  stakingWalletId: {
    type: String,
    required: true,
    trim: true,
  },
  type: {
    type: String,
    required: true,
    trim: true,
  },
  txId: {
    type: String,
    required: true,
  },
  createdDate: {
    type: Date,
    required: true,
  },
  totalReward: {
    type: String,
    required: true,
  },
  amount: {
    type: String,
    required: true,
  },
  initialStakingLimit: {
    type: String,
    required: true,
  },
  claimed: {
    type: String,
    required: true,
  },
});

const Wallets = mongoose.model("StakingAddress", WalletSchema);
module.exports = Wallets;
