const pool_name = (value) => {
	if(value.length > 30){
		return false;
	}
   return /^[a-zA-Z0-9_ ]*$/.test(value);
};
   
module.exports = (fields) => {
    return (req, res, next) => {
    let params = req.body;

    if (req.method === "GET") params = req.params;

    let errors = fields.filter((field) => {
        if (params[field] && !pool_name(params[field].trim())) return field;
    });

    if (errors.length)
        return res.send(422, {
        status: false,
        message: `The ${errors.join(", ")} is not valid`,
        });

    return next();
    };
};
