-- For wallets that have at least one transfer, the current balance_in_cents in
-- mart_wallets must equal the balance_after_in_cents of the most recent transfer.
-- Returns wallets where the two values disagree (expected: 0 rows).
WITH latest_transfer AS (
    SELECT DISTINCT ON (wallet_id)
        wallet_id,
        balance_after_in_cents
    FROM
        {{ ref('mart_transfers') }}
    ORDER BY
        wallet_id,
        transferred_at DESC,
        transfer_id DESC
)
SELECT
    w.wallet_id,
    w.balance_in_cents AS wallet_balance_in_cents,
    lt.balance_after_in_cents AS latest_transfer_balance_after_in_cents
FROM
    {{ ref('mart_wallets') }} w
JOIN
    latest_transfer lt ON lt.wallet_id = w.wallet_id
WHERE
    w.balance_in_cents != lt.balance_after_in_cents
