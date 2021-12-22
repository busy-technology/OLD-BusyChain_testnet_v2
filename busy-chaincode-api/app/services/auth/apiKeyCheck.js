const repository = require("../../repositories/domain/find-domain-by-name-and-key"),
  getDomainType = require("../../helpers/get-domain-type");

module.exports = (req) => {
  return new Promise(async (resolve, reject) => {
    try {
      const doc = await repository({
        domainname: getDomainType(req.headers.apitype),
        apikey: req.headers.apikey,
      });

      return resolve({
        search: "Match found",
      });
    } catch (err) {
      return reject(err);
    }
  });
};
