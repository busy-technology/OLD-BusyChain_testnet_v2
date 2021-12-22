const FabricCAServices = require("fabric-ca-client");
const { Wallets } = require("fabric-network");
const Admin = require("../app/models/admin");
const fs = require("fs");
const path = require("path");

const adminUserId = "busy_network";

exports.saveOrdererAdmin = async () => {
  try {
    const ccpPath = path.resolve(
      __dirname,
      "connection-profile",
      "connection-busy.json"
    );
    const ccp = JSON.parse(fs.readFileSync(ccpPath, "utf8"));

    // Create a new CA client for interacting with the CA.
    const caInfo = ccp.certificateAuthorities["ca.busy.technology"];
    const caTLSCACerts = caInfo.tlsCACerts.pem;
    const ca = new FabricCAServices(
      caInfo.url,
      { trustedRoots: caTLSCACerts, verify: false },
      caInfo.caName
    );
    // Create a new file system based wallet for managing identities.
    const walletPath = path.join(
      process.cwd(),
      "blockchain",
      "network",
      "wallet"
    );
    const wallet = await Wallets.newFileSystemWallet(walletPath);

    // Check to see if we've already enrolled the admin user.
    const identity = await wallet.get(adminUserId);
    const adminData = await Admin.findOne({ userId: adminUserId });
    if (identity && adminData == null) {
      console.log("identity OrdererAdmin", identity);

      const credential = {
        certificate: identity.credentials.certificate,
        privateKey: identity.credentials.privateKey,
      };

      const certificate = {
        credentials: credential,
        mspId: identity.mspId,
        type: identity.type,
        version: "1",
      };

      const admin = await new Admin({
        certificate: certificate,
        userId: adminUserId,
      });

      await admin
        .save()
        .then((result, error) => {
          console.log(" Orderer Admin registered.");
        })
        .catch((error) => {
          console.log("ERROR DB", error);
        });
    } else {
      console.log("Orderer Admin do not exists or already registered in DB.");
    }
  } catch (exception) {
    return exception;
  }
};