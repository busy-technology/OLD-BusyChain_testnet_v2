const mongoose = require("mongoose");
const timeStamp = require("mongoose-timestamp");

const UserSchema = new mongoose.Schema({
  firstName: {
    type: String,
    required: false,
    trim: true,
  },
  lastName: {
    type: String,
    required: false,
    trim: true,
  },
  email: {
    type: String,
    required: false,
    trim: true,
  },
  mobile: {
    type: Number,
    required: false,
    trim: true,
  },
  userId: {
    type: String,
    required: true,
    trim: true,
  },
  walletId: {
    type: String,
    required: false,
  },
  password: {
    type: String,
    required: true,
    trim: true,
  },
  country: {
    type: String,
    required: false,
  },
  txId: {
    type: String,
    required: true,
  },
  tokens: {
    type: Object,
    required: true,
  },
  messageCoins: {
   type: Object,
   required: true
  }
});

UserSchema.plugin(timeStamp);

const User = mongoose.model("BusyUsers", UserSchema);
module.exports = User;
