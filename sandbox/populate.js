const { randomUUID } = require("node:crypto");

const BASE_URL = "http://localhost:8080";
const ACCOUNTS = 25;
const ITERATIONS = 25;

const wallets = {};

async function main() {
	await createWallets();
	await populateWallets();
	await transferBetweenWallets();
}

async function createWallets() {
	for (let i = 0; i < ACCOUNTS; i++) {
		const walletID = randomUUID();
		const holderID = randomUUID();

		await fetch(`${BASE_URL}/wallets`, {
			method: "post",
			body: JSON.stringify({
				wallet_id: walletID,
				holder_id: holderID,
			}),
			headers: { "content-type": "application/json" },
		});

		wallets[walletID] = 0;
		console.log(`Wallet created: ${walletID}`);
	}
}

async function populateWallets() {
	for (const walletID in wallets) {
		const amountInCents = 999_999;
		const transferID = randomUUID();
		const fromWalletID = randomUUID();

		await fetch(`${BASE_URL}/wallets/${walletID}/mock-transfer`, {
			method: "post",
			body: JSON.stringify({
				transfer_id: transferID,
				from_wallet_id: fromWalletID,
				amount_in_cents: amountInCents,
			}),
			headers: { "content-type": "application/json" },
		});

		wallets[walletID] += amountInCents;
		console.log(`Wallet populated: ${walletID}`);
	}
}

async function transferBetweenWallets() {
	for (let i = 0; i < ITERATIONS; i++) {
		const walletIDs = Object.keys(wallets);

		for (let j = 0; j < walletIDs.length; j++) {
			const amountInCents = 100;
			const transferID = randomUUID();
			const fromWalletID = walletIDs[j];
			const toWalletID = walletIDs[(j + 1) % walletIDs.length];

			await fetch(`${BASE_URL}/wallets/${fromWalletID}/transfer`, {
				method: "post",
				body: JSON.stringify({
					transfer_id: transferID,
					to_wallet_id: toWalletID,
					amount_in_cents: amountInCents,
				}),
				headers: { "content-type": "application/json" },
			});

			wallets[fromWalletID] -= amountInCents;
			wallets[toWalletID] += amountInCents;
			console.log(`Transfer: ${fromWalletID} -> ${toWalletID}`);
		}
	}
}

main();
