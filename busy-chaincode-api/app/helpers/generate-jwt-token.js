const jwt = require("jsonwebtoken");
const config = require("../../config");

module.exports = (params) => {
  return jwt.sign(params, config.JWT_SECRET, {
    expiresIn: parseInt(config.EXPIRY_TIME),
  });
};
