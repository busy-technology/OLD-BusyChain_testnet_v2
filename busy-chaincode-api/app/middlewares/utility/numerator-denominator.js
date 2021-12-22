const numerator = (value) => {
    if (/^0.*$/.test(value)) {
        return false
    }
    return /^[0-9]+$/.test(value);
};

const denominator = (value) => {
    if (/^0.*$/.test(value)) {
        return false
    }
    return /^[0-9]+$/.test(value);
};

const fraction = (num, denom) => {
    var numerator = parseInt(num);
    var denominator = parseInt(denom);
    if(numerator/denominator > 1){
        return false;
    }
    return true;
}

module.exports = (fields) => {
    return (req, res, next) => {
        let params = req.body;

        if (req.method === "GET") params = req.params;
        let errors = fields.filter((field) => {
            if (field == fields[0]) {
                if (params[field] && !numerator(params[field].trim())) return field;
            }
            if (field == fields[1]) {
                if (params[field] && !denominator(params[field].trim())) return field;
            }
        });

        if (errors.length)
        return res.send(422, {
            status: false,
            message: `The ${errors.join(", ")} is not valid`,
        });

        if (!fraction(params[fields[0]].trim(),params[fields[1]].trim())){
            return res.send(422, {
                status: false,
                message: `The vesting fraction is not valid`,
            });
        }

        return next();
    };
};