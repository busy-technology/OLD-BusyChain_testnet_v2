const User = require("../../models/Users");
const {
  Certificate
} = require("@fidm/x509");
const bs58 = require("bs58");
const constants = require("../../../constants");
const submitTransaction = require("../../../blockchain/submitTransaction");
const transactions = require("../../models/transactions");
const IssuetokenTransactions = require("../../models/issued-tokens");

module.exports = async (req, res, next) => {
  try {
    const walletId = req.body.walletId,
      blockchain_credentials = req.body.credentials,
      tokenName = req.body.tokenName,
      symbol = req.body.symbol,
      amount = req.body.amount,
      decimals = req.body.decimals;

    console.log("TokenName", tokenName);
    console.log("Symbol", symbol);

    const user = await User.findOne({
      walletId: walletId
    });
    console.log("User", user);

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
      const lowerTokenName = tokenName.toLowerCase();
      console.log("lowerTokenName", lowerTokenName);
      const lowerToken = symbol.toLowerCase();
      console.log("LOWER TOKEN", lowerToken);
      const coinSymbol = await IssuetokenTransactions.findOne({
        symbol: lowerToken,
      });
      console.log("COIN", coinSymbol);
      const coinName = await IssuetokenTransactions.findOne({
        name: lowerTokenName,
      });
      console.log("COIN", coinName);
      if (!coinName) {
        if (!coinSymbol) {
          const decodedPrivateKey = bs58.decode(
            blockchain_credentials.credentials.privateKey
          );

          blockchain_credentials.credentials.privateKey =
            decodedPrivateKey.toString();

          const response = await submitTransaction.SubmitTransaction(
            constants.BUSY_CHANNEL_NAME,
            constants.DEFAULT_CONTRACT_NAME,
            "IssueToken",
            walletId,
            blockchain_credentials,
            tokenName,
            symbol,
            amount,
            decimals
          );
          const resp = JSON.parse(response);

          if (resp.success == true) {
            const tokenEntry = await new IssuetokenTransactions({
              tokenName: tokenName,
              tokenSymbol: resp.data.tokenSymbol,
              transactionId: resp.txId,
              tokenDecimals: decimals,
              createdDate: new Date(),
              tokenAdmin: walletId,
              tokenId: resp.data.id,
              tokenSupply: resp.data.totalSupply,
              logoPath: "",
              websiteUrl: "",
              socialMedia: "",
            });

            await tokenEntry
              .save()
              .then((result, error) => {
                console.log("Issue Token saved.");
              })
              .catch((error) => {
                console.log("ERROR DB", error);
              });
            const transactionEntry = await new transactions({
              transactionType: "issue",
              transactionId: resp.txId,
              submitTime: new Date(),
              payload: {
                name: resp.data.tokenName,
                amount: amount,
                tokenSymbol: resp.data.tokenSymbol,
                symbol: resp.data.tokenSymbol,
                tokenAdmin: resp.data.admin,
                tokenId: resp.data.id,
                tokenSupply: resp.data.totalSupply,
                tokenDecimals: resp.data.decimals,
                sender: "Busy network",
                receiver: walletId,
              },
              status: "submitted"
            });

            await transactionEntry
              .save()
              .then((result, error) => {
                console.log("Issue Tokens transaction recorded.");
              })
              .catch((error) => {
                console.log("ERROR DB", error);
              });


            return res.send(200, {
              status: true,
              message: "Request to create a new token has been successfully accepted",
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
        } else {
          console.log("The symbol is already taken");
          return res.send(404, {
            status: false,
            message: `The symbol is already taken`,
          });
        }
      } else {
        console.log("The name is already taken.");
        return res.send(409, {
          status: false,
          message: `The name is already taken`,
        });
      }
    } else {
      console.log("WalletId do not exists.");
      return res.send(404, {
        status: false,
        message: `Wallet does not exist`,
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