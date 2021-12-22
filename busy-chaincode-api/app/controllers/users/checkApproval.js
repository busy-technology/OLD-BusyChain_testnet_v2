const User = require("../../models/Users");
const transactions = require("../../models/transactions");
const {
    Certificate
} = require("@fidm/x509");
const bs58 = require("bs58");
const constants = require("../../../constants");
const queryTransaction = require("../../../blockchain/queryTransaction");

module.exports = async (req, res, next) => {
    const walletId = req.body.walletId;
    const blockchain_credentials = req.body.credentials;
    const operator = req.body.operator;
    try {
        const user = await User.findOne({
            walletId: walletId
        });
        if (user) {
            const commanName = Certificate.fromPEM(
                Buffer.from(blockchain_credentials.credentials.certificate, "utf-8")
            ).subject.commonName;
            console.log("CN", commanName);
            if (user.userId != commanName) {
                return res.send(404, {
                    status: false,
                    message: `User’s certificate is not valid`,
                });
            }

            if (
                blockchain_credentials.type != "X.509" ||
                blockchain_credentials.mspId != "BusyMSP"
            ) {
                console.log("type of certificate incorrect.");
                return res.send(400, {
                    status: false,
                    message: `Incorrect type or MSPID`,
                });
            }
            const decodedPrivateKey = bs58.decode(
                blockchain_credentials.credentials.privateKey
            );

            blockchain_credentials.credentials.privateKey =
                decodedPrivateKey.toString();

            const response = await queryTransaction.QueryTransaction(
                constants.BUSY_CHANNEL_NAME,
                constants.DEFAULT_CONTRACT_NAME,
                "BusyTokens:IsApprovedForAll",
                walletId,
                blockchain_credentials,
                walletId,
                operator
            );

            const resp = JSON.parse(response);
            if (resp.success == true) {
                return res.send(200, {
                    status: true,
                    message: "Approval status has been successfully fetched",
                    chaincodeResponse: resp,
                })
            } else {
                console.log("Failed to execute chaincode function");
                return res.send(500, {
                    status: false,
                    message: resp.message,
                });
            };
        } else {
            console.log("WalletId do not exists.");
            return res.send(404, {
                status: false,
                message: `Wallet does not exist`,
            });
        }
    } catch (exception) {
        return res.send(500, {
            status: false,
            message: exception.message,
        });
    };
};