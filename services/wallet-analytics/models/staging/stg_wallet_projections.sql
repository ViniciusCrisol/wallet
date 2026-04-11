WITH source AS (
    SELECT *
    FROM {{ source('wallet_service', 'wallet_projections') }}
)
SELECT
    wallet_id,
    holder_id,
    balance_in_cents,
    created_at,
    updated_at
FROM
    source
