const { getRandomValues } = require("node:crypto");

const ACCOUNTS = 10;
const BASE_URL = "http://localhost:8080";

const wallets = new Map();

async function main() {
	await createWallets();
	await populateWallets();
	await transferBetweenWallets();
}

async function createWallets() {
	for (let i = 0; i < ACCOUNTS; i++) {
		const walletID = randomUUID();
		const createdAt = randomDate(new Date("2019-12-01T00:00:00Z"), new Date("2019-12-31T12:00:00Z"));

		await fetch(`${BASE_URL}/wallets`, {
			method: "post",
			body: JSON.stringify({
				wallet_id: walletID,
				holder_id: randomUUID(),
				created_at: createdAt.toISOString(),
			}),
			headers: { "content-type": "application/json" },
		});
		wallets.set(walletID, createdAt);

		console.log(`Wallet created: ${walletID}`);
	}
}

async function populateWallets() {
	for (const [wallet, createdAt] of wallets) {
		const amountInCents = 9_999_999;
		const transferredAt = randomDate(createdAt, new Date(createdAt.getTime() + 30 * 24 * 3600000));

		await fetch(`${BASE_URL}/wallets/${wallet}/mock-transfer`, {
			method: "post",
			body: JSON.stringify({
				transfer_id: randomUUID(),
				from_wallet_id: wallet,
				amount_in_cents: amountInCents,
				transferred_at: transferredAt.toISOString(),
			}),
			headers: { "content-type": "application/json" },
		});
		wallets.set(wallet, transferredAt);

		console.log(`Wallet populated: ${wallet}`);
	}
}

async function transferBetweenWallets() {
	const walletIDs = [...wallets.keys()];
	let i = 0;
	while (true) {
		for (let j = 0; j < walletIDs.length; j++) {
			const fromWalletID = walletIDs[j];

			let toWalletID = walletIDs[Math.floor(Math.random() * walletIDs.length)];
			while (toWalletID === fromWalletID) {
				toWalletID = walletIDs[Math.floor(Math.random() * walletIDs.length)];
			}

			let amountInCents = Math.floor(Math.random() * 10000) + 100;
			if (amountInCents > 1000 && amountInCents < 1010) amountInCents = amountInCents * 100.25;
			if (amountInCents > 2000 && amountInCents < 2100) amountInCents = amountInCents * 50.25;
			if (amountInCents > 3000 && amountInCents < 3999) amountInCents = amountInCents * 5.25;
			amountInCents = amountInCents * Number(`1.${i}`);
			amountInCents = parseInt(amountInCents);

			const lastTransfer =
				wallets.get(fromWalletID) > wallets.get(toWalletID)
					? wallets.get(fromWalletID)
					: wallets.get(toWalletID);
			const transferredAt = randomDate(
				(() => {
					const hoursToAdd = 2;
					return new Date(lastTransfer.getTime() + hoursToAdd * 3600000);
				})(),
				(() => {
					const hoursToAdd = 36;
					return new Date(lastTransfer.getTime() + hoursToAdd * 3600000);
				})(),
			);
			if (transferredAt > new Date()) {
				return;
			}

			const transfer = async () => {
				const response = await fetch(`${BASE_URL}/wallets/${fromWalletID}/transfer`, {
					method: "post",
					body: JSON.stringify({
						transfer_id: randomUUID(),
						to_wallet_id: toWalletID,
						amount_in_cents: amountInCents,
						transferred_at: transferredAt.toISOString(),
					}),
					headers: { "content-type": "application/json" },
				});
				if (response.status === 409) {
					await transfer();
				}
			};
			await transfer();
			wallets.set(fromWalletID, transferredAt);
			wallets.set(toWalletID, transferredAt);

			console.log(`Transfer: ${fromWalletID} -> ${toWalletID}`);
		}
	}
}

function randomDate(min, max) {
	return new Date(min.getTime() + Math.random() * (max.getTime() - min.getTime()));
}

function randomUUID() {
	const timestamp = Date.now();
	const randomValues = getRandomValues(new Uint8Array(10));

	const bytes = new Uint8Array(16);
	bytes[0] = (timestamp / 0x10000000000) & 0xff;
	bytes[1] = (timestamp / 0x100000000) & 0xff;
	bytes[2] = (timestamp / 0x1000000) & 0xff;
	bytes[3] = (timestamp / 0x10000) & 0xff;
	bytes[4] = (timestamp / 0x100) & 0xff;
	bytes[5] = timestamp & 0xff;

	bytes.set(randomValues, 6);
	bytes[6] = (bytes[6] & 0x0f) | 0x70;
	bytes[8] = (bytes[8] & 0x3f) | 0x80;

	return [...bytes]
		.map((b, i) => (i === 4 || i === 6 || i === 8 || i === 10 ? "-" : "") + b.toString(16).padStart(2, "0"))
		.join("");
}

main();
