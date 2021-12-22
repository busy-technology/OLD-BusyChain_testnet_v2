const Admin = require("../../models/admin");
const queryTransaction = require("../../../blockchain/queryTransaction");
const constants = require("../../../constants");

module.exports = async (req, res, next) => {
    try {
        const tokenSymbol = req.query.tokenSymbol,
            adminId = "busy_network";

        const adminData = await Admin.findOne({
            userId: adminId
        });

        const credentials = {
            certificate: adminData.certificate.credentials.certificate,
            privateKey: adminData.certificate.credentials.privateKey,
        };

        const blockchain_credentials = {
            credentials: credentials,
            mspId: adminData.certificate.mspId,
            type: adminData.certificate.type,
        };


        if(!tokenSymbol){
            return res.send(400, {
                status: false,
                message: `Empty Symbol in the request`,
            });
        }

        const response = await queryTransaction.QueryTransaction(
            constants.BUSY_CHANNEL_NAME,
            constants.DEFAULT_CONTRACT_NAME,
            "BusyTokens:GetTokenInfo",
            adminId,
            blockchain_credentials,
            tokenSymbol,
        );

        console.log(response);
        const resp = JSON.parse(response);
        if (resp.success == true) {
            return res.send(200, {
                status: true,
                message: "The token information has been successfully fetched",
                chaincodeResponse: resp,
            });
        } else {
            console.log("Failed to execute chaincode function");
            return res.send(500, {
                status: false,
                message: `Failed to execute chaincode function`,
                chaincodeResponse: resp,
            });
        }
    } catch (exception) {
        console.log(exception);
        return res.send(500, {
            status: false,
            message: exception.message,
        });
    }
};