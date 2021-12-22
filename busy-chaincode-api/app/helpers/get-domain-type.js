module.exports = (apitype) => {
  return apitype
    .replace(/^(?:http?:\/\/)?(?:https?:\/\/)?(?:www\.)?/i, "")
    .split("/")[0];
};
