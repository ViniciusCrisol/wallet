WITH source AS (
    SELECT *
    FROM {{ source('wallet_service', 'transfer_projections') }}
)
SELECT
    wallet_id,
    transfer_id,
    counterpart_wallet_id,
    direction,
    category,
    amount_in_cents,
    transferred_at
FROM
    source
