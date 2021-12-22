const is_password = (value) => {
  return true
  // supports internation nammes
};

module.exports = (fields) => {
  return (req, res, next) => {
    let params = req.body;

    if (req.method === "GET") params = req.params;

    let errors = fields.filter((field) => {
      if (params[field] && !is_password(params[field].trim())) return field;
    });

    if (errors.length)
      return res.send(422, {
        status: false,
        message: `The ${errors.join(", ")} is not valid`,
      });

    return next();
  };
};
