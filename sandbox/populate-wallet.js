const { randomUUID } = require("node:crypto");

const BASE_URL = "http://localhost:8080";
const ACCOUNTS = 50;
const ITERATIONS = 100;

const wallets = [];

async function main() {
	await createWallets();
	await populateWallets();
	await transferBetweenWallets();
}

async function createWallets() {
	for (let i = 0; i < ACCOUNTS; i++) {
		const walletID = randomUUID();

		await fetch(`${BASE_URL}/wallets`, {
			method: "post",
			body: JSON.stringify({
				wallet_id: walletID,
				holder_id: randomUUID(),
			}),
			headers: { "content-type": "application/json" },
		});

		wallets.push(walletID);
		console.log(`Wallet created: ${walletID}`);
	}
}

async function populateWallets() {
	for (const wallet in wallets) {
		const amountInCents = 999_999;

		await fetch(`${BASE_URL}/wallets/${wallet}/mock-transfer`, {
			method: "post",
			body: JSON.stringify({
				transfer_id: randomUUID(),
				from_wallet_id: randomUUID(),
				amount_in_cents: amountInCents,
			}),
			headers: { "content-type": "application/json" },
		});

		console.log(`Wallet populated: ${wallet}`);
	}
}

async function transferBetweenWallets() {
	for (let i = 0; i < ITERATIONS; i++) {
		for (let j = 0; j < wallets.length; j++) {
			const fromWalletID = wallets[j];

			const amountInCents = Math.floor(Math.random() * 199) + 1;

			let toWalletID = wallets[Math.floor(Math.random() * wallets.length)];
			while (toWalletID === fromWalletID) {
				toWalletID = wallets[Math.floor(Math.random() * wallets.length)];
			}

			const transfer = async () => {
				const response = await fetch(`${BASE_URL}/wallets/${fromWalletID}/transfer`, {
					method: "post",
					body: JSON.stringify({
						transfer_id: randomUUID(),
						to_wallet_id: toWalletID,
						amount_in_cents: amountInCents,
					}),
					headers: { "content-type": "application/json" },
				});
				if (response.status === 409) {
					await transfer();
				}
			};
			await transfer();

			console.log(`Transfer: ${fromWalletID} -> ${toWalletID}`);
		}
	}
}

main();
