DELETE FROM accounts WHERE is_system = TRUE;  -- clean Settlement
DROP TABLE IF EXISTS accounts;