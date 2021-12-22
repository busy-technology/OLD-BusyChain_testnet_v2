/*const service = require("../../services/auth/generate-token-admin");

module.exports = (req, res, next) => {
  service(req)
    .then((response) => {
      return res.send(200, {
        status: true,
        message: "Authentication token has been successfully generated",
        data: response,
      });
    })
    .catch((error) => {
      return res.send(error.code, { status: false, message: error.message });
    })
    .finally(() => {
      return next();
    });
};
*/