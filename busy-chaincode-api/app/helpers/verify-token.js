const jwt = require("jsonwebtoken");
const config = require("../../config");

module.exports = (token) => {
  return jwt.verify(token, config.JWT_SECRET);
};
