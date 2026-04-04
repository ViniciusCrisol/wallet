CREATE TABLE wallet.wallet_projections (
	wallet_id        VARCHAR(36) NOT NULL PRIMARY KEY,
	holder_id        VARCHAR(36) NOT NULL,
	balance_in_cents INT         NOT NULL,
	created_at       DATETIME(6) NOT NULL,
	updated_at       DATETIME(6) NOT NULL
);

CREATE TABLE wallet.transfer_projections (
	wallet_id             VARCHAR(36) NOT NULL,
	transfer_id           VARCHAR(36) NOT NULL,
	counterpart_wallet_id VARCHAR(36) NOT NULL,
	direction             VARCHAR(10) NOT NULL,
	category              VARCHAR(20) NOT NULL,
	amount_in_cents       INT         NOT NULL,
	transferred_at        DATETIME(6) NOT NULL,
	PRIMARY KEY (wallet_id, transfer_id),
	CONSTRAINT fk_transfer_wallet_id             FOREIGN KEY (wallet_id)             REFERENCES wallet_projections(wallet_id),
	CONSTRAINT fk_transfer_counterpart_wallet_id FOREIGN KEY (counterpart_wallet_id) REFERENCES wallet_projections(wallet_id)
);
CREATE INDEX idx_wallet_transfers ON wallet.transfer_projections (wallet_id, transferred_at DESC);
