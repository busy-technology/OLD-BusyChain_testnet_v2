const Admin = require("../../models/admin");
const submitTransaction = require("../../../blockchain/submitTransaction");
const constants = require("../../../constants");
const transactions = require("../../models/transactions");
const queryTransaction = require("../../../blockchain/queryTransaction");

module.exports = {
    updateMessagingFee: async (req, res, next) => {
        try {
            const newFee = req.body.newFee;
            const adminId = "busy_network";
            const userId = "sample";

            console.log("IN USER");
            const adminData = await Admin.findOne({
                userId: adminId
            });
            console.log("ADMIN", adminData);

            const credentials = {
                certificate: adminData.certificate.credentials.certificate,
                privateKey: adminData.certificate.credentials.privateKey
            };

            const blockchain_credentials = {
                credentials: credentials,
                mspId: adminData.certificate.mspId,
                type: adminData.certificate.type
            };

            const response = await submitTransaction.SubmitTransaction(
                    constants.BUSY_CHANNEL_NAME,
                    constants.DEFAULT_CONTRACT_NAME,
                    "BusyMessenger:UpdateMessagingFee",
                    userId,
                    blockchain_credentials,
                    newFee
                    );
            const resp = JSON.parse(response);

            if (resp.success == true) {
                const transactionEntry = await new transactions({
                    transactionType: "updateMessagingFee",
                    transactionId: resp.txId,
                    submitTime: new Date(),
                    payload: {
                        newFee: newFee
                    },
                    status: "submitted"
                });

                await transactionEntry
                        .save()
                        .then((result, error) => {
                            console.log("Messaging Fees transaction recorded.");
                        })
                        .catch((error) => {
                            console.log("ERROR DB", error);
                        });

                return res.send(200, {
                    status: true,
                    message: "Request to update transaction fee has been successfully accepted",
                    chaincodeResponse: resp
                });
            } else {
                console.log("Failed to execute chaincode function");
                return res.send(500, {
                    status: false,
                    message: `Failed to execute chaincode function`,
                    chaincodeResponse: resp
                });
            }
        } catch (exception) {
            console.log(exception);
            return res.send(500, {
                status: false,
                message: exception.message
            });
        }
    },
    updateTokenIssueFee: async (req, res, next) => {
        try {
            const newFee = req.body.newFee;
            const tokenType = req.body.tokenType;

            const adminId = "busy_network";
            const userId = "sample";

            const adminData = await Admin.findOne({
                userId: adminId
            });
            console.log("ADMIN", adminData);

            const credentials = {
                certificate: adminData.certificate.credentials.certificate,
                privateKey: adminData.certificate.credentials.privateKey
            };

            const blockchain_credentials = {
                credentials: credentials,
                mspId: adminData.certificate.mspId,
                type: adminData.certificate.type
            };

            const response = await submitTransaction.SubmitTransaction(
                    constants.BUSY_CHANNEL_NAME,
                    constants.DEFAULT_CONTRACT_NAME,
                    "SetTokenIssueFee",
                    userId,
                    blockchain_credentials,
                    tokenType,
                    newFee
                    );
            const resp = JSON.parse(response);

            if (resp.success == true) {
                const transactionEntry = await new transactions({
                    transactionType: "UpdateTokenIssueFee",
                    transactionId: resp.txId,
                    submitTime: new Date(),
                    payload: {
                        newFee: newFee,
                        tokenType: tokenType
                    },
                    status: "submitted"
                });

                await transactionEntry
                        .save()
                        .then((result, error) => {
                            console.log("Update Token Issue Fee transaction recorded.");
                        })
                        .catch((error) => {
                            console.log("ERROR DB", error);
                        });

                return res.send(200, {
                    status: true,
                    message: "Request to update transaction fee has been successfully accepted",
                    chaincodeResponse: resp
                });
            } else {
                console.log("Failed to execute chaincode function");
                return res.send(500, {
                    status: false,
                    message: `Failed to execute chaincode function`,
                    chaincodeResponse: resp
                });
            }
        } catch (exception) {
            console.log(exception);
            return res.send(500, {
                status: false,
                message: exception.message
            });
        }
    },
    getMessagingFee: async (req, res, next) => {
        const adminId = "busy_network";
        const adminData = await Admin.findOne({userId: adminId});

        const credentials = {
            certificate: adminData.certificate.credentials.certificate,
            privateKey: adminData.certificate.credentials.privateKey
        };

        const blockchain_credentials = {
            credentials: credentials,
            mspId: adminData.certificate.mspId,
            type: adminData.certificate.type
        };

        try {
            const response = await queryTransaction.QueryTransaction(
                    constants.BUSY_CHANNEL_NAME,
                    constants.DEFAULT_CONTRACT_NAME,
                    "BusyMessenger:GetMessagingFee",
                    adminId,
                    blockchain_credentials
                    );

            const resp = JSON.parse(response);
            if (resp.success == true) {
                console.log("Current messaging fee has been successfully fetched");
                return res.send(200, {
                    status: true,
                    message: "Current messaging fee has been successfully fetched",
                    chaincodeResponse: resp
                });
            } else {
                console.log("Failed to execute chaincode function");
                return res.send(500, {
                    status: false,
                    message: resp.message
                });
            }
        } catch (exception) {
            console.log(exception);
            return res.send(500, {
                status: false,
                message: exception.message
            });
        }
    },
    getTokenIssueFee: async (req, res, next) => {
        const adminId = "busy_network";
        const adminData = await Admin.findOne({userId: adminId});

        const credentials = {
            certificate: adminData.certificate.credentials.certificate,
            privateKey: adminData.certificate.credentials.privateKey
        };

        const blockchain_credentials = {
            credentials: credentials,
            mspId: adminData.certificate.mspId,
            type: adminData.certificate.type
        };

        try {
            const response = await queryTransaction.QueryTransaction(
                    constants.BUSY_CHANNEL_NAME,
                    constants.DEFAULT_CONTRACT_NAME,
                    "GetTokenIssueFee",
                    adminId,
                    blockchain_credentials
                    );

            const resp = JSON.parse(response);
            if (resp.success == true) {
                console.log("Current token issue fee has been successfully fetched");
                return res.send(200, {
                    status: true,
                    message: "Current token issue fee fee has been successfully fetched",
                    chaincodeResponse: resp
                });
            } else {
                console.log("Failed to execute chaincode function");
                return res.send(500, {
                    status: false,
                    message: resp.message
                });
            }
        } catch (exception) {
            console.log(exception);
            return res.send(500, {
                status: false,
                message: exception.message
            });
        }
    }
};