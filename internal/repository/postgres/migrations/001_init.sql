-- 001_init.sql

-- Optional but recommended for UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- -------------------------
-- merchants
-- -------------------------
CREATE TABLE IF NOT EXISTS merchants (
  id              BIGSERIAL PRIMARY KEY,
  merchant_uid    UUID NOT NULL DEFAULT gen_random_uuid(), -- public ID
  name            TEXT NOT NULL,
  wallet_address  TEXT NOT NULL,
  status          TEXT NOT NULL DEFAULT 'PENDING',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT merchants_status_check
    CHECK (status IN ('PENDING', 'ACTIVE', 'INACTIVE'))
);

-- If merchant name must be unique globally (as you said):
CREATE UNIQUE INDEX IF NOT EXISTS merchants_name_uidx ON merchants (name);

-- Public ID unique:
CREATE UNIQUE INDEX IF NOT EXISTS merchants_merchant_uid_uidx ON merchants (merchant_uid);

-- Often wallet should be unique too (optional):
CREATE UNIQUE INDEX IF NOT EXISTS merchants_wallet_uidx ON merchants (wallet_address);

-- -------------------------
-- users (login accounts)
-- -------------------------
CREATE TABLE IF NOT EXISTS users (
  id              BIGSERIAL PRIMARY KEY,
  user_uid        UUID NOT NULL DEFAULT gen_random_uuid(),
  merchant_id     BIGINT REFERENCES merchants(id) ON DELETE SET NULL, -- admin can be NULL
  email           TEXT NOT NULL,
  password_hash   TEXT NOT NULL,
  role            TEXT NOT NULL DEFAULT 'MERCHANT', -- ADMIN / MERCHANT / OPERATOR
  status          TEXT NOT NULL DEFAULT 'ACTIVE',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT users_role_check
    CHECK (role IN ('ADMIN', 'MERCHANT', 'OPERATOR')),
  CONSTRAINT users_status_check
    CHECK (status IN ('ACTIVE', 'DISABLED'))
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_uidx ON users (email);
CREATE UNIQUE INDEX IF NOT EXISTS users_user_uid_uidx ON users (user_uid);

-- -------------------------
-- orders
-- -------------------------
CREATE TABLE IF NOT EXISTS orders (
  id              BIGSERIAL PRIMARY KEY,
  order_uid        UUID NOT NULL DEFAULT gen_random_uuid(), -- public ID
  merchant_id      BIGINT NOT NULL REFERENCES merchants(id),
  order_id         TEXT NOT NULL,     -- your external order id (string)
  invoice_id       TEXT NOT NULL,     -- your invoice id (string)
  amount           NUMERIC(36,18) NOT NULL,
  currency         TEXT NOT NULL DEFAULT 'USDT',
  token_address    TEXT,              -- for TRC20 token contract address if needed
  payment_status   TEXT NOT NULL DEFAULT 'PENDING',
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT orders_payment_status_check
    CHECK (payment_status IN ('PENDING', 'SUCCESS', 'FAILED'))
);

-- Your “unique” requirements (better per merchant):
CREATE UNIQUE INDEX IF NOT EXISTS orders_order_id_uidx ON orders (merchant_id, order_id);
CREATE UNIQUE INDEX IF NOT EXISTS orders_invoice_id_uidx ON orders (merchant_id, invoice_id);
CREATE UNIQUE INDEX IF NOT EXISTS orders_order_uid_uidx ON orders (order_uid);

CREATE INDEX IF NOT EXISTS orders_merchant_idx ON orders (merchant_id);
CREATE INDEX IF NOT EXISTS orders_status_idx ON orders (payment_status);

-- -------------------------
-- payments
-- -------------------------
CREATE TABLE IF NOT EXISTS payments (
  id               BIGSERIAL PRIMARY KEY,
  payment_uid      UUID NOT NULL DEFAULT gen_random_uuid(), -- public ID
  order_id         BIGINT NOT NULL REFERENCES orders(id),
  merchant_id      BIGINT NOT NULL REFERENCES merchants(id),

  invoice_id       TEXT NOT NULL,
  token_address    TEXT,
  payer_address    TEXT,
  merchant_address TEXT,

  amount           NUMERIC(36,18) NOT NULL,
  currency         TEXT NOT NULL DEFAULT 'USDT',

  tx_hash          TEXT,
  status           TEXT NOT NULL DEFAULT 'PENDING',
  confirmed_at     TIMESTAMPTZ,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT payments_status_check
    CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED'))
);

CREATE UNIQUE INDEX IF NOT EXISTS payments_payment_uid_uidx ON payments (payment_uid);

-- tx_hash can be NULL for not-yet-known, but once present should be unique:
CREATE UNIQUE INDEX IF NOT EXISTS payments_tx_hash_uidx ON payments (tx_hash) WHERE tx_hash IS NOT NULL;

CREATE INDEX IF NOT EXISTS payments_order_idx ON payments (order_id);
CREATE INDEX IF NOT EXISTS payments_merchant_idx ON payments (merchant_id);
CREATE INDEX IF NOT EXISTS payments_status_idx ON payments (status);
