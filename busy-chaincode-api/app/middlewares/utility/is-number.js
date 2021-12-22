const is_number = (value) => {
  return /^\+[1-9]{1}[0-9]{3,14}$/.test(value);
// + followed by your mobile number
  ///^-?\d+$/
};

module.exports = (fields) => {
  return (req, res, next) => {
    let params = req.body;

    if (req.method === "GET") params = req.params;

    let errors = fields.filter((field) => {
      if (params[field] && !is_number(params[field].trim())) return field;
    });

    if (errors.length)
      return res.send(422, {
        status: false,
        message: `The ${errors.join(", ")} is not valid`,
      });

    return next();
  };
};
