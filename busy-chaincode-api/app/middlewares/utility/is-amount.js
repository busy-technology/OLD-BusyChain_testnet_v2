const is_amount = (value) => {
  if(/^0.*$/.test(value)){
    return false
  }
  return /^[0-9]+$/.test(value);
  //allows space and letter only
};

module.exports = (fields) => {
  return (req, res, next) => {
    let params = req.body;

    if (req.method === "GET") params = req.params;

    let errors = fields.filter((field) => {
      
      
      if (params[field] && (typeof(params[field])=='string') &&(params[field].length>0) && !is_amount(params[field].trim())) return field;
    });

    if (errors.length)
      return res.send(422, {
        status: false,
        message: `The ${errors.join(", ")} is not valid`,
      });

    return next();
  };
};
