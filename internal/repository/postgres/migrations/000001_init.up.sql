-- =====================================================
-- 001_init.sql
-- Token13 Merchant Backend
-- Canonical IDs = bytes32 (BYTEA)
-- =====================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =====================================================
-- merchants
-- =====================================================
CREATE TABLE IF NOT EXISTS merchants (
  merchant_id          BYTEA PRIMARY KEY,       -- bytes32 (canonical ID)

  name                 TEXT NOT NULL,
  wallet_address       TEXT NOT NULL,

  status               TEXT NOT NULL DEFAULT 'PENDING',

  chain_txid           TEXT,
  chain_registered_at  TIMESTAMPTZ,

  created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT merchants_status_check
    CHECK (status IN ('PENDING','ACTIVE','INACTIVE'))
);

CREATE UNIQUE INDEX IF NOT EXISTS merchants_wallet_uidx
  ON merchants (wallet_address);

CREATE UNIQUE INDEX IF NOT EXISTS merchants_name_uidx
  ON merchants (name);

-- =====================================================
-- users
-- =====================================================
CREATE TABLE IF NOT EXISTS users (
  id               BIGSERIAL PRIMARY KEY,

  user_uid         UUID NOT NULL DEFAULT gen_random_uuid(),

  -- NULL for platform admins
  merchant_id      BYTEA REFERENCES merchants(merchant_id) ON DELETE SET NULL,

  email            TEXT NOT NULL,
  password_hash    TEXT NOT NULL,

  role             TEXT NOT NULL DEFAULT 'MERCHANT',
  status           TEXT NOT NULL DEFAULT 'ACTIVE',

  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT users_role_check
    CHECK (role IN ('ADMIN','MERCHANT','OPERATOR')),

  CONSTRAINT users_status_check
    CHECK (status IN ('ACTIVE','DISABLED'))
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_uidx
  ON users (email);

CREATE UNIQUE INDEX IF NOT EXISTS users_user_uid_uidx
  ON users (user_uid);

-- =====================================================
-- orders
-- =====================================================
CREATE TABLE IF NOT EXISTS orders (
  id                 BIGSERIAL PRIMARY KEY,

  order_id           BYTEA NOT NULL,           -- bytes32 (canonical)
  invoice_id         BYTEA NOT NULL,           -- bytes32 (canonical)

  merchant_id        BYTEA NOT NULL REFERENCES merchants(merchant_id),

  amount             NUMERIC(36,18) NOT NULL,
  currency           TEXT NOT NULL DEFAULT 'USDT',

  token_address      TEXT,
  payment_status     TEXT NOT NULL DEFAULT 'PENDING',

  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT orders_payment_status_check
    CHECK (payment_status IN ('PENDING','SUCCESS','FAILED'))
);

-- Canonical uniqueness
CREATE UNIQUE INDEX IF NOT EXISTS orders_order_id_uidx
  ON orders (order_id);

CREATE UNIQUE INDEX IF NOT EXISTS orders_invoice_id_uidx
  ON orders (invoice_id);

CREATE INDEX IF NOT EXISTS orders_merchant_idx
  ON orders (merchant_id);

CREATE INDEX IF NOT EXISTS orders_status_idx
  ON orders (payment_status);

-- =====================================================
-- payments
-- =====================================================
CREATE TABLE IF NOT EXISTS payments (
  id                  BIGSERIAL PRIMARY KEY,

  payment_uid         UUID NOT NULL DEFAULT gen_random_uuid(),

  order_id            BYTEA NOT NULL REFERENCES orders(order_id),
  invoice_id          BYTEA NOT NULL,

  merchant_id         BYTEA NOT NULL REFERENCES merchants(merchant_id),

  token_address       TEXT,
  payer_address       TEXT,
  merchant_address    TEXT,

  amount              NUMERIC(36,18) NOT NULL,
  currency            TEXT NOT NULL DEFAULT 'USDT',

  tx_hash             TEXT,
  status              TEXT NOT NULL DEFAULT 'PENDING',

  confirmed_at        TIMESTAMPTZ,

  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT payments_status_check
    CHECK (status IN ('PENDING','SUCCESS','FAILED'))
);

CREATE UNIQUE INDEX IF NOT EXISTS payments_payment_uid_uidx
  ON payments (payment_uid);

CREATE UNIQUE INDEX IF NOT EXISTS payments_tx_hash_uidx
  ON payments (tx_hash)
  WHERE tx_hash IS NOT NULL;

CREATE INDEX IF NOT EXISTS payments_order_idx
  ON payments (order_id);

CREATE INDEX IF NOT EXISTS payments_merchant_idx
  ON payments (merchant_id);

CREATE INDEX IF NOT EXISTS payments_status_idx
  ON payments (status);
