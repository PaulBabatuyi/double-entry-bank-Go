CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    balance NUMERIC(19,4) NOT NULL DEFAULT 0.0000,
    currency TEXT NOT NULL DEFAULT 'NGN',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed the system Settlement account (external cash flow)
INSERT INTO accounts (id, name, balance, currency, is_system)
VALUES (
    gen_random_uuid(),
    'Settlement Account',
    0.0000,
    'NGN',
    TRUE
);