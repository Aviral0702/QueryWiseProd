const express = require('express');
const axios = require('axios');
const jwt = require('jsonwebtoken');
const crypto = require('crypto');
const { pool } = require('../db');
const { requireAuth, JWT_SECRET } = require('../middleware/auth');

const router = express.Router();

const GOOGLE_CLIENT_ID = process.env.GOOGLE_CLIENT_ID;
const GOOGLE_CLIENT_SECRET = process.env.GOOGLE_CLIENT_SECRET;
const FRONTEND_URL = (process.env.FRONTEND_URL || 'http://localhost:3001').replace(/\/$/, '');
const BACKEND_URL = (process.env.BACKEND_URL || process.env.API_URL || `http://localhost:${process.env.PORT || 3000}`).replace(/\/$/, '');

if (!GOOGLE_CLIENT_ID || !GOOGLE_CLIENT_SECRET) {
  console.warn('QueryWise: GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set for Google SSO. Auth routes will return 503.');
}

// GET /api/v1/auth/google — redirect to Google OAuth
router.get('/google', (req, res) => {
  if (!GOOGLE_CLIENT_ID) {
    return res.status(503).json({ error: 'Google SSO is not configured' });
  }
  const state = crypto.randomBytes(16).toString('hex');
  const redirectUri = `${BACKEND_URL}/api/v1/auth/google/callback`;
  const scope = encodeURIComponent('openid email profile');
  const url = `https://accounts.google.com/o/oauth2/v2/auth?response_type=code&client_id=${encodeURIComponent(GOOGLE_CLIENT_ID)}&redirect_uri=${encodeURIComponent(redirectUri)}&scope=${scope}&state=${state}&access_type=offline&prompt=consent`;
  res.redirect(url);
});

// GET /api/v1/auth/google/callback — exchange code, create/find user, issue JWT, redirect to frontend
router.get('/google/callback', async (req, res) => {
  if (!GOOGLE_CLIENT_ID || !GOOGLE_CLIENT_SECRET) {
    return res.redirect(`${FRONTEND_URL}/login?error=config`);
  }
  const { code, state, error: oauthError } = req.query;
  if (oauthError) {
    return res.redirect(`${FRONTEND_URL}/login?error=${encodeURIComponent(oauthError)}`);
  }
  if (!code) {
    return res.redirect(`${FRONTEND_URL}/login?error=no_code`);
  }

  const redirectUri = `${BACKEND_URL}/api/v1/auth/google/callback`;

  try {
    const tokenRes = await axios.post(
      'https://oauth2.googleapis.com/token',
      new URLSearchParams({
        code,
        client_id: GOOGLE_CLIENT_ID,
        client_secret: GOOGLE_CLIENT_SECRET,
        redirect_uri: redirectUri,
        grant_type: 'authorization_code',
      }),
      { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
    );
    const { id_token, access_token } = tokenRes.data;

    const userRes = await axios.get('https://www.googleapis.com/oauth2/v2/userinfo', {
      headers: { Authorization: `Bearer ${access_token}` },
    });
    const { id: google_id, email, name, picture } = userRes.data;

    const r = await pool.query(
      `INSERT INTO users (google_id, email, name, avatar_url)
       VALUES ($1, $2, $3, $4)
       ON CONFLICT (google_id) DO UPDATE SET email = $2, name = $3, avatar_url = $4
       RETURNING id, email, name, avatar_url`,
      [google_id, email || '', name || '', picture || null]
    );
    const user = r.rows[0];

    const token = jwt.sign(
      {
        sub: String(user.id),
        email: user.email,
        name: user.name,
      },
      JWT_SECRET,
      { expiresIn: '7d' }
    );

    // Redirect to frontend with token in fragment so it is not sent to server logs
    res.redirect(`${FRONTEND_URL}/auth/callback#token=${encodeURIComponent(token)}`);
  } catch (err) {
    console.error('Google OAuth callback error:', err.response?.data || err.message);
    res.redirect(`${FRONTEND_URL}/login?error=callback`);
  }
});

// GET /api/v1/auth/me — return current user (requires Authorization: Bearer <token>)
router.get('/me', requireAuth, (req, res) => {
  res.json({ user: req.user });
});

module.exports = router;
