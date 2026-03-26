require('dotenv').config();

const express = require('express');
const cors = require('cors');
const { requireApiKey, requireAuth } = require('./middleware/auth');
const { ingestLimiter } = require('./middleware/rateLimit');
const ingestRouter = require('./routes/ingest');
const reportsRouter = require('./routes/reports');
const instancesRouter = require('./routes/instances');
const authRouter = require('./routes/auth');
const rateLimit = require('express-rate-limit');

const app = express();

const FRONTEND_URL = (process.env.FRONTEND_URL || 'http://localhost:3001').replace(/\/$/, '');
const allowedOrigins = [FRONTEND_URL];
(process.env.FRONTEND_URLS || '')
  .split(',')
  .map((s) => s.trim().replace(/\/$/, ''))
  .filter(Boolean)
  .forEach((o) => {
    if (!allowedOrigins.includes(o)) allowedOrigins.push(o);
  });
if (FRONTEND_URL === 'http://localhost:3001') {
  allowedOrigins.push('http://127.0.0.1:3001');
}

// Basic security headers (lightweight alternative to adding Helmet as a dependency)
app.disable('x-powered-by');
app.use((req, res, next) => {
  res.setHeader('X-Content-Type-Options', 'nosniff');
  res.setHeader('X-Frame-Options', 'DENY');
  next();
});

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

app.use(express.json({ limit: '1mb' }));

// Lightweight request logging in production
app.use((req, res, next) => {
  res.on('finish', () => {
    if (process.env.NODE_ENV === 'production') {
      console.log(`${req.method} ${req.originalUrl} -> ${res.statusCode}`);
    }
  });
  next();
});

// Health
app.get('/api/health', (req, res) => {
  res.json({ status: 'ok', service: 'QueryWise Backend' });
});

app.get('/api/v1/health', (req, res) => {
  res.json({ status: 'ok', service: 'QueryWise Backend' });
});

// Auth: Google OAuth (no auth) + /me (requireAuth applied in router)
app.use(
  '/api/v1/auth',
  // Prevent brute-force against OAuth-related endpoints / misconfigured users
  rateLimit({ windowMs: 60 * 1000, max: 30, message: { error: 'Too many requests' } }),
  authRouter
);

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

// 404 + error handlers (consistent JSON errors for UI)
app.use((req, res) => {
  res.status(404).json({ error: 'Not found' });
});

// eslint-disable-next-line no-unused-vars
app.use((err, req, res, next) => {
  console.error('Unhandled error:', err);
  res.status(500).json({ error: 'Internal server error' });
});

app.listen(PORT, () => {
  console.log(`QueryWise Backend listening on port ${PORT}`);
});
