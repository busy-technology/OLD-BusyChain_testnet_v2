module.exports = {
  ENV: process.env.NODE_ENV || "development",
  PORT: process.env.PORT || "3000",
  URL: process.env.BASE_URL || "http://localhost:3000",
  MONGODB_URI:
    process.env.MONGODB_URI ||
    "mongodb://127.0.0.1:27017/busy?authSource=admin",
  EXPIRY_TIME: process.env.EXPIRY_TIME || "60",
  JWT_SECRET: process.env.JWT_SECRET || "BUSY SOLUIONS ARE ALWAYS PROTECTED",
};
