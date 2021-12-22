module.exports = {
  utility: {
    required: require("./utility/required"),
    number: require("./utility/is-number"),
    userId: require("./utility/is-userId"),
    isEmail: require("./utility/is-email"),
    isPassword: require("./utility/is-password"),
    isCountry: require("./utility/is-country"),
    isName: require("./utility/is-name"),
    isAmount: require("./utility/is-amount"),
    isAlphaNumeric: require("./utility/is-alphanumeric"),
    voteType: require("./utility/vote-type"),
    isPoolName: require("./utility/pool-name"),
    isPoolDescription: require("./utility/pool-description"),
    isTime: require("./utility/is-time"),
    isNumeratorDenominator: require("./utility/numerator-denominator"),
    isNumeric: require("./utility/is-numeric"),
    isArray: require("./utility/is-array"),
    isAmounts: require("./utility/is-amounts"),
  },
  auth: {
    generateToken: require("./auth/generate-token"),
  },
};
