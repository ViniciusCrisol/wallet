WITH wallets AS (
    SELECT *
    FROM {{ ref('stg_wallet_projections') }}
),
transfer_stats AS (
    SELECT
        wallet_id,
        COUNT(1) AS total_transfers,
        MAX(transferred_at) AS last_transfer_at,
        MAX(amount_in_cents) AS largest_transfer_in_cents,
        SUM(amount_in_cents) AS total_transferred_amount_in_cents,
        COUNT(1) FILTER (WHERE transferred_at >= now() - INTERVAL '30 days') AS transfers_last_30_days,
        COUNT(1) FILTER (WHERE transferred_at >= now() - INTERVAL '90 days') AS transfers_last_90_days
    FROM
        {{ ref('stg_transfer_projections') }}
    GROUP BY
        wallet_id
),
favorite_category AS (
    SELECT DISTINCT ON (wallet_id)
        wallet_id,
        category AS favorite_category
    FROM
        (
            SELECT
                wallet_id,
                category,
                COUNT(1) AS category_transfer_count
            FROM
                {{ ref('stg_transfer_projections') }}
            GROUP BY
                wallet_id,
                category
        )
    ORDER BY
        wallet_id,
        category_transfer_count DESC
)
SELECT
    w.wallet_id,
    w.holder_id,
    w.balance_in_cents,
    fc.favorite_category,
    CASE
        WHEN COALESCE(ts.last_transfer_at, w.created_at) >= now() - INTERVAL '90 days'
        THEN 'active'
        ELSE 'suspended'
    END AS status,
    ts.last_transfer_at,
    COALESCE(ts.total_transfers, 0) AS total_transfers,
    COALESCE(ts.transfers_last_30_days, 0) AS transfers_last_30_days,
    COALESCE(ts.transfers_last_90_days, 0) AS transfers_last_90_days,
    COALESCE(ts.largest_transfer_in_cents, 0) AS largest_transfer_in_cents,
    COALESCE(ts.total_transferred_amount_in_cents, 0) AS total_transferred_amount_in_cents,
    w.created_at,
    w.updated_at
FROM
    wallets w
LEFT JOIN
    transfer_stats ts ON ts.wallet_id = w.wallet_id
LEFT JOIN
    favorite_category fc ON fc.wallet_id = w.wallet_id
