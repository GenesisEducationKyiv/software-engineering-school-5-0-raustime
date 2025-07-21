-- Create subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR NOT NULL UNIQUE,
    city VARCHAR NOT NULL,
    frequency VARCHAR NOT NULL,
    confirmed BOOLEAN NOT NULL DEFAULT false,
    token VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    confirmed_at TIMESTAMPTZ
);

-- Create index for better performance
CREATE INDEX IF NOT EXISTS idx_subscriptions_email ON subscriptions(email);
CREATE INDEX IF NOT EXISTS idx_subscriptions_confirmed ON subscriptions(confirmed);
CREATE INDEX IF NOT EXISTS idx_subscriptions_city ON subscriptions(city);