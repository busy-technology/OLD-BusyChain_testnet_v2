const is_voting_time = (value) => {
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
        if (params[field] && !is_voting_time(params[field])) return field;
      });
  
      if (errors.length)
        return res.send(422, {
          status: false,
          message: `The ${errors.join(", ")} is not valid`,
        });
  
      return next();
    };
  };
  