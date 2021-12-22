/*const repository = require("../../repositories/domain/find-admin-domain-and-key"),
  getDomainType = require("../../helpers/get-domain-type"),
  uuid = require("uuid-random"),
  generateToken = require("../../helpers/generate-jwt-token");

module.exports = (req) => {
  return new Promise(async (resolve, reject) => {
    try {
      if (
        req.headers.apitype == "busy.admins" &&
        req.headers.apikey == "hckch874867487njkbjvw89797"
      ) {
        const doc = await repository({
          domainname: getDomainType(req.headers.apitype),
          apikey: req.headers.apikey,
        });

        return resolve({
          token: generateToken({
            _id: doc._id,
            domainname: doc.domainname,
            uuid: uuid(),
          }),
        });
      } else {
        console.log("IN ELSE");
        return reject({
          code: 404,
          message: "Domain incorrect for admin access.",
        });
      }
    } catch (err) {
      return reject(err);
    }
  });
};
*/