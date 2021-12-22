const vote_type = (value) => {
   if(value != "Yes" && value != "No"){
       return false;
   }
   return true;
  };
  
  module.exports = (fields) => {
    return (req, res, next) => {
      let params = req.body;
  
      if (req.method === "GET") params = req.params;
  
      let errors = fields.filter((field) => {
        if (params[field] && !vote_type(params[field].trim())) return field;
      });
  
      if (errors.length)
        return res.send(422, {
          status: false,
          message: `The ${errors.join(", ")} is not valid`,
        });
  
      return next();
    };
  };
  