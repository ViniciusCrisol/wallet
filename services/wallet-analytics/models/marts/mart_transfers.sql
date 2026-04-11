WITH transfers AS (
    SELECT *
    FROM {{ ref('stg_transfer_projections') }}
),
wallets AS (
    SELECT
        wallet_id,
        balance_in_cents
    FROM
        {{ ref('stg_wallet_projections') }}
),
transfers_with_balance AS (
    SELECT
        t.wallet_id,
        t.transfer_id,
        t.counterpart_wallet_id,
        t.category,
        t.direction,
        t.amount_in_cents,
        t.transferred_at,
        CASE t.direction
            WHEN 'incoming'
            THEN t.amount_in_cents
            ELSE - t.amount_in_cents
        END AS signed_amount_in_cents,
        w.balance_in_cents AS current_balance_in_cents
    FROM
        transfers t
    JOIN
        wallets w ON w.wallet_id = t.wallet_id
)
SELECT
    wallet_id,
    transfer_id,
    counterpart_wallet_id,
    category,
    direction,
    amount_in_cents,
    current_balance_in_cents - suffix_sum_in_cents AS balance_before_in_cents,
    current_balance_in_cents - suffix_sum_in_cents + signed_amount_in_cents AS balance_after_in_cents,
    transferred_at
FROM
    (
        SELECT
            *,
            SUM(signed_amount_in_cents) OVER (
                PARTITION BY wallet_id
                ORDER BY transferred_at DESC, transfer_id DESC
                ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
            ) AS suffix_sum_in_cents
        FROM
            transfers_with_balance
    )
