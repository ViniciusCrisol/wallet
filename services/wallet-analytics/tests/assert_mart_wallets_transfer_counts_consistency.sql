-- Transfer count windows must form a monotone hierarchy:
--   transfers_last_30_days <=
--   transfers_last_90_days <=
--   total_transfers
-- Returns wallets that violate this ordering (expected: 0 rows).
SELECT
    wallet_id,
    total_transfers,
    transfers_last_90_days,
    transfers_last_30_days
FROM
    {{ ref('mart_wallets') }}
WHERE
    transfers_last_30_days > transfers_last_90_days OR transfers_last_90_days > total_transfers
