const domains = require("../../models/domain");
const constants = require("../../../constants");

module.exports = async () => {
    domains.find({}).exec(async function (err, result) {
        if (err) {
            throw new Error(err);
        }
        if (result.length == 0) {
            const domainEntryUser = await new domains({
                domainname: constants.DOMAIN_NAME_USER,
                apikey: constants.API_KEY_USER
            });

            await domainEntryUser
                    .save()
                    .then((result, error) => {
                        console.log("Inserting USER domains into the db");
                    })
                    .catch((error) => {
                        console.log("ERROR DB", error);
                    });

            const domainEntryAdmin = await new domains({
                domainname: constants.DOMAIN_NAME_ADMIN,
                apikey: constants.API_KEY_ADMIN
            });

            await domainEntryAdmin
                    .save()
                    .then((result, error) => {
                        console.log("Inserting Admin domains into the db");
                    })
                    .catch((error) => {
                        console.log("Error")
                        throw new Error(error);
                    });

            const domainEntryWallet = await new domains({
                domainname: constants.DOMAIN_NAME_WALLET,
                apikey: constants.API_KEY_WALLET
            });


            await domainEntryWallet
                    .save()
                    .then((result, error) => {
                        console.log("Inserting WALLET domains into the db");
                    })
                    .catch((error) => {
                        console.log("ERROR DB", error);
                    });

            const domainEntryTestnet = await new domains({
                domainname: constants.DOMAIN_NAME_TESTNET,
                apikey: constants.API_KEY_TESTNET
            });


            await domainEntryTestnet
                    .save()
                    .then((result, error) => {
                        console.log("Inserting TESTNET domains into the db");
                    })
                    .catch((error) => {
                        console.log("ERROR DB", error);
                    });
        } else {
            console.log("Domains already exists in the db");
        }
    });

};