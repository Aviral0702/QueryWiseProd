const jwt = require('jsonwebtoken');
const { pool } = require('../db');

const JWT_SECRET = process.env.JWT_SECRET || process.env.API_SECRET || 'change-me-in-production';

async function requireApiKey(req, res, next) {
  const apiKey = req.headers['x-api-key'] || req.query.api_key;
  if (!apiKey) {
    return res.status(401).json({ error: 'Missing X-API-Key' });
  }
  const r = await pool.query(
    'SELECT id, name FROM db_instances WHERE api_key = $1',
    [apiKey]
  );
  if (r.rows.length === 0) {
    return res.status(401).json({ error: 'Invalid API key' });
  }
  req.dbInstance = r.rows[0];
  next();
}

function requireAuth(req, res, next) {
  const authHeader = req.headers.authorization;
  const token = authHeader && authHeader.startsWith('Bearer ') ? authHeader.slice(7) : null;
  if (!token) {
    return res.status(401).json({ error: 'Authentication required' });
  }
  try {
    if (JWT_SECRET === 'change-me-in-production' || !JWT_SECRET) {
      console.error('QueryWise: JWT secret not configured (set JWT_SECRET or API_SECRET).');
      return res.status(500).json({ error: 'Server authentication is not configured' });
    }
    const payload = jwt.verify(token, JWT_SECRET);
    req.user = { id: payload.sub, email: payload.email, name: payload.name };
    next();
  } catch (err) {
    return res.status(401).json({ error: 'Invalid or expired token' });
  }
}

module.exports = { requireApiKey, requireAuth, JWT_SECRET };
