-- Every transfer must have a strictly positive amount.
-- Returns rows where amount_in_cents is zero or negative (expected: 0 rows).
SELECT
    wallet_id,
    transfer_id,
    amount_in_cents
FROM
    {{ ref('mart_transfers') }}
WHERE
    amount_in_cents <= 0
