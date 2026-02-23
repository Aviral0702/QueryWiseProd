const rateLimit = require('express-rate-limit');

const ingestLimiter = rateLimit({
  windowMs: 60 * 1000,
  max: 120,
  message: { error: 'Too many requests' },
  standardHeaders: true,
  legacyHeaders: false,
});

module.exports = { ingestLimiter };
