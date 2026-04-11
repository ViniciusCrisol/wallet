CREATE TABLE wallet_projections (
	wallet_id        VARCHAR(36)  NOT NULL PRIMARY KEY,
	holder_id        VARCHAR(36)  NOT NULL,
	balance_in_cents INT          NOT NULL,
	created_at       TIMESTAMP(6) NOT NULL,
	updated_at       TIMESTAMP(6) NOT NULL
);

CREATE TABLE transfer_projections (
	wallet_id             VARCHAR(36)                 NOT NULL,
	transfer_id           VARCHAR(36)                 NOT NULL,
	counterpart_wallet_id VARCHAR(36)                 NOT NULL,
	category              VARCHAR(20)                 NOT NULL,
	direction             VARCHAR(10)                 NOT NULL,
	amount_in_cents       INT                         NOT NULL,
	transferred_at        TIMESTAMP(6) WITH TIME ZONE NOT NULL,
	PRIMARY KEY (wallet_id, transfer_id)
);
CREATE INDEX idx_wallet_transfers ON transfer_projections (wallet_id, transferred_at DESC);
