const User = require("../../models/Users");
const bcrypt = require("bcrypt");

module.exports = async (req, res, next) => {
  const userId = req.body.userId,
    password = req.body.password;

  User.findOne({ userId: userId })
    .then((user) => {
      console.log("USER DATA", user);
      const result = bcrypt.compareSync(password, user.password);
      if (result == false) {
        console.log("Password is incorrect.");
        return res.send(400, {
          status: false,
          message: "Password is incorrect",
        });
      } else {
        return res.send(200, {
          status: true,
          message: "Login has been successful",
        });
      }
    })
    .catch((error) => {
      console.log("User data not found in users table.");
      if (error) {
        console.log("error", error);
        return res.send(404, {
          status: false,
          message: "User does not exist",
        });
      }
    });
};
