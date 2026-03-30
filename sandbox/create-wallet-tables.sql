CREATE TABLE wallet.wallet_projections (
	wallet_id        VARCHAR(36) NOT NULL PRIMARY KEY,
	holder_id        VARCHAR(36) NOT NULL,
	balance_in_cents INT         NOT NULL,
	created_at       DATETIME(6) NOT NULL,
	updated_at       DATETIME(6) NOT NULL
);
