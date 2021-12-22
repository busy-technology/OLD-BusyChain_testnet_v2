let trim_req = (params) => {
  // chek if params is not empy/undefined
  if (params !== null && typeof params === "object") {
    for (let key in params) {
      // check if property is object in itself
      if (typeof params[key] === "object") return trim_req(params[key]);

      // elseif value is string
      if (typeof params[key] === "string") params[key] = params[key].trim();
    }
  }
};

module.exports = trim_req;
