const transactions = require("../../models/transactions");
module.exports = async (req, res, next) => {
    var startBlock = req.query.startBlock;
    var endBlock = req.query.endBlock;

    if (!req.query.walletId) {
        var query = await transactions.find({
            blockNum: {
                $gt: startBlock, $lt: endBlock
            }
        });
    } else {
        var walletId = req.query.walletId;
        console.log(walletId);

        var query = await transactions.find({
            $and: [

                {blockNum: {$gt: startBlock, $lt: endBlock}},
                {"payload.address": walletId}
            ]
        });
    }

    if (!query) {
        return res.send(500, {
            status: false,
            message: "Error occured while fetching the transactions",
        });
    }
    return res.send(200, {
        status: true,
        message: "Transactions have been successfully fetched",
        data: query,
    });
};