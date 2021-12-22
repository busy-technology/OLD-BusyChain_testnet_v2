const verifyToken = require("../helpers/verify-token");
const checkJwt = require("./checkJwt");

module.exports = async (req, res, next) => {
    const errors = ["authorization"].filter(
            (field) => !req.headers[field]
    );

    if (errors.length)
        return res.send(401, {
            status: false,
            message: `Header properties required: ${errors.join(", ")}`
        });

    try {
        token = verifyToken(req.headers.authorization);
    } catch (err) {
        return res.send(403, {
            status: false,
            message: `Token Error: ${err.message}`
        });
    }
    
    let httpUrl = new URL("http://" + req.headers.host + req.url);
    var authResp = checkJwt(req.headers.authorization, httpUrl.pathname.substring(1));
    if (!authResp.authorized) {
        return res.send(authResp.statusCode, {
            status: false,
            message: authResp.errorMsg
        });
    }
    return next();
};