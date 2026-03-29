CREATE TABLE wallet.wallet_projections (
	wallet_id        VARCHAR(36) NOT NULL PRIMARY KEY,
	holder_id        VARCHAR(36) NOT NULL,
	balance_in_cents INT         NOT NULL,
	created_at       DATETIME(6) NOT NULL,
	updated_at       DATETIME(6) NOT NULL
);

CREATE TABLE wallet.processed_events (
	event_id     VARCHAR(255) NOT NULL,
	event_type   VARCHAR(255) NOT NULL,
	processed_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (event_id, event_type)
);
