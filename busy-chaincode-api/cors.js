module.exports = (req, res, next) => {
    // Website you wish to allow to connect
    res.setHeader("Access-Control-Allow-Origin", "*");
  
    // Request methods you wish to allow
    res.setHeader(
      "Access-Control-Allow-Methods",
      "GET, POST, PUT, PATCH, DELETE"
    );
  
    // Request headers you wish to allow
    res.setHeader(
      "Access-Control-Allow-Headers",
      "authorization, apikey, apitype, accesstoken, userid, uuid,Content-Type,x-access-token"
    );
  
    // Set to true if you need the website to include cookies in the requests sent
    // to the API (e.g. in case you use sessions)
    res.setHeader("Access-Control-Allow-Credentials", false);
  
    if (req.method === "OPTIONS")
      // if is preflight(OPTIONS) then response status 204(NO CONTENT)
      return res.send(204);
  
    return next();
  };
  