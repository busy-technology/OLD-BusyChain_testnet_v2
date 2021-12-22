const Admin = require("../../models/admin");
module.exports = async (req, res, next) => {
  try {
    //const userId = "admin";
    const userId = "busy_network";
    const credentials = req.body.credentials;

    const credential = {
      certificate: credentials.credentials.certificate,
      privateKey: credentials.credentials.privateKey,
    };

    const certificate = {
      credentials: credential,
      mspId: credentials.mspId,
      type: credentials.type,
      version: "1",
    };

    const admin = await new Admin({
      certificate: certificate,
      userId: userId,
    });

    await admin
      .save()
      .then((result, error) => {
        console.log("Admin registered.");
      })
      .catch((error) => {
        console.log("ERROR DB", error);
      });

    return res.send(200, {
      status: true,
      message: "Admin registered.",
    });
  } catch (exception) {
    console.log(exception);
    return res.send(500, {
      status: false,
      message: exception.message,
    });
  }
};
