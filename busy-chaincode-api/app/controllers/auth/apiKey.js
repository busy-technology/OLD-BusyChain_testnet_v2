const service = require("../../services/auth/apiKeyCheck");

module.exports = (req, res, next) => {
  service(req)
    .then((response) => {
      //   return res.send(200, {
      //     status: true,
      //     message: "API key is valid.",
      //     data: response,
      //   });
      console.log("API key is valid");
      next();
    })
    .catch((error) => {
      return res.send(error.code, { status: false, message: error.message });
    });
  // .finally(() => {
  //   return next();
  // });
};
