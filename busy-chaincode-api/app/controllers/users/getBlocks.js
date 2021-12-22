const blocks = require("../../models/blocks");

module.exports = async (req, res, next) => {
  var query = await blocks.find({})
  if (!query) {
    return res.send(500, {
      status: false,
      message: "Error occured while fetching the blocks",
    });
  }
  return res.send(200, {
    status: true,
    message: "Blocks have been successfully fetched",
    data: query,
  });
};