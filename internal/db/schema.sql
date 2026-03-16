-- deposits 充值记录表
CREATE TABLE IF NOT EXISTS deposits (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    tx_id       TEXT NOT NULL UNIQUE,
    address     TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    amount      REAL NOT NULL,
    height      INTEGER NOT NULL,
    confirmed   INTEGER NOT NULL DEFAULT 0,
    chain       TEXT NOT NULL CHECK(chain IN ('btc','eth')),
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- withdrawals 提币记录表（后面会用到）
CREATE TABLE IF NOT EXISTS withdrawals (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    tx_id       TEXT,
    address     TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    amount      REAL NOT NULL,
    fee         REAL NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'pending',
    chain       TEXT NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- deposit_addresses 充值地址表
CREATE TABLE IF NOT EXISTS deposit_addresses (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    TEXT NOT NULL,
    address    TEXT NOT NULL UNIQUE,
    chain      TEXT NOT NULL CHECK(chain IN ('btc','eth')),
    path       TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);