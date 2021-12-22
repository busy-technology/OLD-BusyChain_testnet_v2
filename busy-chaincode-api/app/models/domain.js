const mongoose = require("mongoose");
const timeStamp = require("mongoose-timestamp");

const domainSchema = new mongoose.Schema({
  domainname: {
    type: String,
    required: true,
  },
  apikey: {
    type: String,
    required: true,
  },
});

domainSchema.plugin(timeStamp);

const Domain = mongoose.model("BusyDomains", domainSchema);
module.exports = Domain;
