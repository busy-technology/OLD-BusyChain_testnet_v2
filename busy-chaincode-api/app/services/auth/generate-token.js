const repository = require("../../repositories/domain/find-domain-by-name-and-key"),
        getDomainType = require("../../helpers/get-domain-type"),
        uuid = require("uuid-random"),
        generateToken = require("../../helpers/generate-jwt-token");

module.exports = (req) => {
    return new Promise(async (resolve, reject) => {
        try {
            //TO BE DONE - FETCH APIKEY WHETHER EXISTS IN THE DB, IF YES, return ACCESS GROUP
            if (req.headers.apikey == "a1b2c33d4e5f6g7h8i9jakblc" || req.headers.apikey == "hckch874867487njkbjvw89797" || req.headers.apikey == "sdla98878pomndakdl97h993" || req.headers.apikey == "dff09mjfdp08djkmsdADDF") {

                let accessGroup = "";
                // BELOW accessGroup will be loaded from Mongo
                if (req.headers.apikey == "hckch874867487njkbjvw89797") {
                    accessGroup = "busyadmin";
                } else if (req.headers.apikey == "sdla98878pomndakdl97h993") {
                    accessGroup = "busywallet";
                } else if (req.headers.apikey == "dff09mjfdp08djkmsdADDF") {
                    accessGroup = "busytestnet";
                } else {
                    accessGroup = "busyuser";
                }

                const doc = await repository({
                    domainname: getDomainType(accessGroup),
                    apikey: req.headers.apikey
                });

                return resolve({
                    token: generateToken({
                        _id: doc._id,
                        domainname: doc.domainname,
                        uuid: uuid()
                    })
                });
            } else {
                console.log("IN ELSE");
                return reject({
                    code: 404,
                    message: "Domain incorrect for User access."
                });
            }
        } catch (err) {
            return reject(err);
        }
    });
};
