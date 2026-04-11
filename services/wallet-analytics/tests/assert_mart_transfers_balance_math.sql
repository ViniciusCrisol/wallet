-- Every transfer must satisfy:
--   balance_after_in_cents = balance_before_in_cents + signed_amount
-- where signed_amount = +amount for incoming, -amount for outgoing.
-- Returns rows that violate this invariant (expected: 0 rows).
SELECT
    wallet_id,
    transfer_id,
    direction,
    amount_in_cents,
    balance_before_in_cents,
    balance_after_in_cents,
    balance_before_in_cents +
    CASE direction
        WHEN 'incoming'
        THEN amount_in_cents
        ELSE - amount_in_cents
    END AS expected_balance_after_in_cents
FROM
    {{ ref('mart_transfers') }}
WHERE
    balance_after_in_cents != balance_before_in_cents +
    CASE direction
        WHEN 'incoming'
        THEN amount_in_cents
        ELSE - amount_in_cents
    END
