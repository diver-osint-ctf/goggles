import { getWebsites, initConfig } from "./sheets";
import type { Config } from "./config";

function sendToDiscord(text: string, config: Config) {
	const content = {
		embeds: [
			{
				title: ":goggles: Health Check by Goggles",
				description: text,
				color: 15158332,
			},
		],
	};
	const options: GoogleAppsScript.URL_Fetch.URLFetchRequestOptions = {
		method: "post",
		headers: {
			"Content-Type": "application/json",
		},
		payload: JSON.stringify(content),
	};
	UrlFetchApp.fetch(config.discordWebhookUrl, options);
}

export function healthCheck() {
	const config = initConfig();
	if (config === null) {
		console.error("config is null");
		return;
	}

	const websites = getWebsites();
	if (websites.length === 0) {
		return;
	}

	const text = `テストです\n\n[debug]\n${websites
		.map((website) => {
			website.healthCheck();
			return JSON.stringify(website.result);
		})
		.join("\n")}`;

	sendToDiscord(text, config);
}
