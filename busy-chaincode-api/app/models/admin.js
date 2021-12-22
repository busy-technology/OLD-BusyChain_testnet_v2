const mongoose = require("mongoose");
const timeStamp = require("mongoose-timestamp");

const credentialSchema = mongoose.Schema({
  certificate: String,
  privateKey: String,
});

const certificateSchema = mongoose.Schema({
  credentials: {
    type: credentialSchema,
  },
  mspId: String,
  type: String,
  version: String,
});

const adminSchema = mongoose.Schema({
  certificate: certificateSchema,
  userId: String,
});

adminSchema.plugin(timeStamp);

const Admin = mongoose.model("Admins", adminSchema);
module.exports = Admin;
