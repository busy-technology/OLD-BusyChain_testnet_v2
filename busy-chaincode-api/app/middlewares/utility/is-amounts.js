const is_amounts = (amounts) => {
    if(amounts == null || amounts == "undefined" || amounts == [] || amounts.length == 0){
        return false;
    } else {
        for(let i = 0;i < amounts.length;i++){
            if(isNaN(amounts[i])){
                return false;
            }
            if(amounts[i] === 0){
                return false;
            }
        }
    }
    return true;
};
  
module.exports = (fields) => {
    return (req, res, next) => {
      let params = req.body;
  
      if (req.method === "GET") params = req.params;
  
      let errors = fields.filter((field) => {
        if (params[field] && !is_amounts(params[field])) return field;
      });
  
      if (errors.length)
        return res.send(422, {
          status: false,
          message: `The ${errors.join(", ")} is not valid`,
        });
  
      return next();
    };
  };
  