CREATE TYPE operation_type AS ENUM ('deposit', 'withdrawal', 'transfer');

CREATE TABLE entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    debit NUMERIC(19,4) NOT NULL DEFAULT 0.0000 CHECK (debit >= 0),
    credit NUMERIC(19,4) NOT NULL DEFAULT 0.0000 CHECK (credit >= 0),
    transaction_id UUID NOT NULL,
    operation_type operation_type NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_single_side CHECK (
        (debit > 0 AND credit = 0) OR (debit = 0 AND credit > 0)
    )
);

-- Optional index for fast lookups (common in ledgers)
CREATE INDEX idx_entries_transaction_id ON entries(transaction_id);
CREATE INDEX idx_entries_account_id ON entries(account_id);