const express = require('express');
const { pool } = require('../db');
const crypto = require('crypto');

const router = express.Router();

function authUserId(req) {
  if (!req.user?.id) return null;
  const id = parseInt(req.user.id, 10);
  return Number.isNaN(id) ? null : id;
}

router.post('/', async (req, res) => {
  const { name } = req.body || {};
  if (!name || typeof name !== 'string') {
    return res.status(400).json({ error: 'Missing "name"' });
  }
  const trimmed = name.trim();
  if (trimmed.length < 2 || trimmed.length > 120) {
    return res.status(400).json({ error: 'Instance name must be 2-120 characters' });
  }
  const uid = authUserId(req);
  const apiKey = crypto.randomBytes(24).toString('hex');
  const r = await pool.query(
    `INSERT INTO db_instances (name, api_key, owner_user_id)
     VALUES ($1, $2, $3)
     RETURNING id, name, api_key, created_at`,
    [trimmed, apiKey, uid]
  );
  res.status(201).json(r.rows[0]);
});

router.get('/', async (req, res) => {
  const uid = authUserId(req);
  const r = uid != null
    ? await pool.query(
        `SELECT i.id, i.name, i.created_at,
                (SELECT MAX(recorded_at) FROM query_metrics m WHERE m.db_instance_id = i.id) AS last_updated
         FROM db_instances i
         WHERE i.owner_user_id = $1
         ORDER BY i.created_at DESC`,
        [uid]
      )
    : await pool.query(
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
  const uid = authUserId(req);
  const reveal = req.query.reveal === '1' || req.query.reveal === 'true';
  const cols = reveal ? 'id, name, created_at, api_key, owner_user_id' : 'id, name, created_at, owner_user_id';
  const r =
    uid != null
      ? await pool.query(`SELECT ${cols} FROM db_instances WHERE id = $1 AND owner_user_id = $2`, [id, uid])
      : await pool.query(`SELECT ${cols} FROM db_instances WHERE id = $1`, [id]);
  if (r.rows.length === 0) return res.status(404).json({ error: 'Instance not found' });
  const row = { ...r.rows[0] };
  delete row.owner_user_id;
  res.json(row);
});

router.post('/:id/regenerate-key', async (req, res) => {
  const id = parseInt(req.params.id, 10);
  if (Number.isNaN(id)) return res.status(400).json({ error: 'Invalid id' });
  const uid = authUserId(req);
  const apiKey = crypto.randomBytes(24).toString('hex');
  const r =
    uid != null
      ? await pool.query(
          'UPDATE db_instances SET api_key = $1 WHERE id = $2 AND owner_user_id = $3 RETURNING id, name, api_key, created_at',
          [apiKey, id, uid]
        )
      : await pool.query(
          'UPDATE db_instances SET api_key = $1 WHERE id = $2 RETURNING id, name, api_key, created_at',
          [apiKey, id]
        );
  if (r.rows.length === 0) return res.status(404).json({ error: 'Instance not found' });
  res.json(r.rows[0]);
});

router.delete('/:id', async (req, res) => {
  const id = parseInt(req.params.id, 10);
  if (Number.isNaN(id)) return res.status(400).json({ error: 'Invalid id' });
  const uid = authUserId(req);
  const r =
    uid != null
      ? await pool.query('DELETE FROM db_instances WHERE id = $1 AND owner_user_id = $2 RETURNING id', [id, uid])
      : await pool.query('DELETE FROM db_instances WHERE id = $1 RETURNING id', [id]);
  if (r.rows.length === 0) return res.status(404).json({ error: 'Instance not found' });
  res.status(204).send();
});

module.exports = router;
