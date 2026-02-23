require('dotenv').config();

const express = require('express');
const cors = require('cors');
const bodyParser = require('body-parser');
const { requireApiKey, requireAuth } = require('./middleware/auth');
const { ingestLimiter } = require('./middleware/rateLimit');
const ingestRouter = require('./routes/ingest');
const reportsRouter = require('./routes/reports');
const instancesRouter = require('./routes/instances');
const authRouter = require('./routes/auth');

const app = express();

const FRONTEND_URL = (process.env.FRONTEND_URL || 'http://localhost:3001').replace(/\/$/, '');
const allowedOrigins = [FRONTEND_URL];
if (FRONTEND_URL === 'http://localhost:3001') {
  allowedOrigins.push('http://127.0.0.1:3001');
}

app.use(cors({
  origin: (origin, cb) => {
    if (!origin) return cb(null, allowedOrigins[0]);
    if (allowedOrigins.includes(origin)) return cb(null, origin);
    cb(null, false);
  },
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization'],
}));
app.use(bodyParser.json());

// Health
app.get('/api/health', (req, res) => {
  res.json({ status: 'ok', service: 'QueryWise Backend' });
});

app.get('/api/v1/health', (req, res) => {
  res.json({ status: 'ok', service: 'QueryWise Backend' });
});

// Auth: Google OAuth (no auth) + /me (requireAuth applied in router)
app.use('/api/v1/auth', authRouter);

const requireGoogleAuth = process.env.ENABLE_GOOGLE_AUTH !== 'false';

if (requireGoogleAuth) {
  app.use('/api/v1/instances', requireAuth, instancesRouter);
  app.use('/api/v1/reports', requireAuth, reportsRouter);
} else {
  app.use('/api/v1/instances', instancesRouter);
  app.use('/api/v1/reports', reportsRouter);
}

// Ingest: agent auth + rate limit (unchanged)
app.use('/api/v1/ingest', requireApiKey, ingestLimiter, ingestRouter);

const PORT = process.env.PORT || 3000;

app.listen(PORT, () => {
  console.log(`QueryWise Backend listening on port ${PORT}`);
});
