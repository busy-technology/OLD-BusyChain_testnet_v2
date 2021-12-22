const FabricCAServices = require("fabric-ca-client");
const { Wallets } = require("fabric-network");
// const { Wallets } = require('fabric-network');
// const FabricCAServices = require('fabric-ca-client');
const fs = require("fs");
const path = require("path");

const adminUserId = "admin";
const adminUserPasswd = process.env.ADMINPW || "adminpw";

exports.FabricAdminEnroll = async () => {
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
    // const walletPath = path.resolve(__dirname, '..', '..', 'network', 'wallet')
    const wallet = await Wallets.newFileSystemWallet(walletPath);
    // const wallet = await new FileSystemWallet(walletPath);

    // Check to see if we've already enrolled the admin user.
    const identity = await wallet.get(adminUserId);
    if (identity) {
      console.log(
        'An identity for the admin user "admin" already exists in the wallet'
      );
      // return {
      //   msg: 'An identity for the admin user "admin" already exists in the wallet',
      // };
    } else {
      // Enroll the admin user, and import the new identity into the wallet.
      const enrollment = await ca.enroll({
        enrollmentID: adminUserId,
        enrollmentSecret: adminUserPasswd,
      });

      // const adminIdentity = X509WalletMixin.createIdentity('akcessMSP', enrollment.certificate, enrollment.key.toBytes())
      // await wallet.import('admin', adminIdentity);

      const x509Identity = {
        credentials: {
          certificate: enrollment.certificate,
          privateKey: enrollment.key.toBytes(),
        },
        mspId: "BusyMSP",
        type: "X.509",
      };
      await wallet.put(adminUserId, x509Identity);
      console.log(
        'Successfully enrolled admin user "admin" and imported it into the wallet'
      );
      return x509Identity;
    }
  } catch (exception) {
    // logger.error(exception.errors);
    return exception;
  }
};
