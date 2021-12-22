const trim_req = require("../../helpers/trim-obj");

module.exports = (req, res, next) => {
  // check for params in body
  if (req.body) trim_req(req.body);

  // check for params in route-url-param
  if (req.params) trim_req(req.params);

  // check for params in query-param
  if (req.query) trim_req(req.query);

  return next();
};
