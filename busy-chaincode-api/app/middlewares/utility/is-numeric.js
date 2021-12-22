const is_numeric = (value) => {
    return /^\d+$/.test(value);
  // + followed by your mobile number
    ///^-?\d+$/
  };
  
module.exports = (fields) => {
    return (req, res, next) => {
      let params = req.body;
  
      if (req.method === "GET") params = req.params;
  
      let errors = fields.filter((field) => {
        if (params[field] && !is_numeric(params[field])) return field;
      });
  
      if (errors.length)
        return res.send(422, {
          status: false,
          message: `The ${errors.join(", ")} is not valid`,
        });
  
      return next();
    };
};
  