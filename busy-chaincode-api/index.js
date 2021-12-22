const mongoose = require("mongoose");
const config = require("./config");
const xss = require("xss-clean");
const helmet = require("helmet");
const mongoSanitize = require("express-mongo-sanitize");
const enrollAdmin = require("./blockchain/enrollAdmin");
const enrollOrdererAdmin = require("./blockchain/enrollOrdererAdmin");
const AdminDb = require("./blockchain/saveAdmin");
const OrdererAdminDb = require("./blockchain/saveOrdererAdmin");
const SaveDomains = require("./app/controllers/users/insertDomains");
const restify = require("restify"),
  server = restify.createServer({
    name: "Busy chaincode API",
    version: "1.0.0",
  }),
  cors = require("./cors");

server.pre(cors);

server.use(restify.plugins.throttle({ burst: 100, rate: 20, ip: true }));

server.use(
  restify.plugins.bodyParser({
    mapParams: false,
    maxBodySize: 1024 * 1024 * 2,
    // requestBodyOnGet: true,
    urlencoded: { extended: false },
  })
);

//server.use(restify.json({ limit: "10kb" })); // body limit is 10

server.use(xss());
server.use(helmet());
server.use(mongoSanitize());

server.use(restify.plugins.queryParser({ mapParams: false }));

server.listen(config.PORT, async () => {
  mongoose.connect(config.MONGODB_URI, {
    useNewUrlParser: true,
    useUnifiedTopology: true,
  });
  await enrollAdmin.FabricAdminEnroll();
  await enrollOrdererAdmin.FabricAdminEnroll();
  await AdminDb.saveAdmin();
  await OrdererAdminDb.saveOrdererAdmin();
  await SaveDomains();
});

const db = mongoose.connection;

db.on("error", (err) => {
  console.log(err);
});

db.once("open", () => {
  require("./app/routes")(server);
  console.log(`Server started on port ${config.PORT}`);
});
