-- 加 email 字段，允许 NULL 方便旧数据过渡
ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT;

-- 旧数据用 username 填充 email（临时兼容）
UPDATE users SET email = username WHERE email IS NULL;

-- 加非空 + 唯一约束
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);
