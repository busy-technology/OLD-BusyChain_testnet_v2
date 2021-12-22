const mongoose = require("mongoose");
const timeStamp = require("mongoose-timestamp");

const PoolSchema = new mongoose.Schema({
  PoolID: {
    type: String,
    required: false,
    trim: true,
  },
  PoolInfo: {
    type: Object,
    requred: true,
  }
});

PoolSchema.plugin(timeStamp);

const Pool = mongoose.model("BusyVotingPool", PoolSchema);
module.exports = Pool;
