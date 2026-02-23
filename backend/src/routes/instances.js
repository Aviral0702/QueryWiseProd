const express = require('express');
const { pool } = require('../db');
const crypto = require('crypto');

const router = express.Router();

// Optional: simple API to register a db_instance and get an api_key (for setup)
// In production you'd protect this or create instances via admin UI.
router.post('/', async (req, res) => {
  const { name } = req.body || {};
  if (!name || typeof name !== 'string') {
    return res.status(400).json({ error: 'Missing "name"' });
  }
  const apiKey = crypto.randomBytes(24).toString('hex');
  const r = await pool.query(
    'INSERT INTO db_instances (name, api_key) VALUES ($1, $2) RETURNING id, name, api_key, created_at',
    [name.trim(), apiKey]
  );
  res.status(201).json(r.rows[0]);
});

router.get('/', async (req, res) => {
  const r = await pool.query(
    `SELECT i.id, i.name, i.created_at,
            (SELECT MAX(recorded_at) FROM query_metrics m WHERE m.db_instance_id = i.id) AS last_updated
     FROM db_instances i
     ORDER BY i.created_at DESC`
  );
  const instances = r.rows.map((row) => ({
    id: row.id,
    name: row.name,
    created_at: row.created_at,
    last_updated: row.last_updated?.toISOString?.() ?? null,
  }));
  res.json({ instances });
});

router.get('/:id', async (req, res) => {
  const id = parseInt(req.params.id, 10);
  if (Number.isNaN(id)) return res.status(400).json({ error: 'Invalid id' });
  const reveal = req.query.reveal === '1' || req.query.reveal === 'true';
  const cols = reveal ? 'id, name, created_at, api_key' : 'id, name, created_at';
  const r = await pool.query(
    `SELECT ${cols} FROM db_instances WHERE id = $1`,
    [id]
  );
  if (r.rows.length === 0) return res.status(404).json({ error: 'Instance not found' });
  res.json(r.rows[0]);
});

router.post('/:id/regenerate-key', async (req, res) => {
  const id = parseInt(req.params.id, 10);
  if (Number.isNaN(id)) return res.status(400).json({ error: 'Invalid id' });
  const apiKey = crypto.randomBytes(24).toString('hex');
  const r = await pool.query(
    'UPDATE db_instances SET api_key = $1 WHERE id = $2 RETURNING id, name, api_key, created_at',
    [apiKey, id]
  );
  if (r.rows.length === 0) return res.status(404).json({ error: 'Instance not found' });
  res.json(r.rows[0]);
});

router.delete('/:id', async (req, res) => {
  const id = parseInt(req.params.id, 10);
  if (Number.isNaN(id)) return res.status(400).json({ error: 'Invalid id' });
  const r = await pool.query('DELETE FROM db_instances WHERE id = $1 RETURNING id', [id]);
  if (r.rows.length === 0) return res.status(404).json({ error: 'Instance not found' });
  res.status(204).send();
});

module.exports = router;
