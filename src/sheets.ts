import Website from "./Website";
import type { Config } from "./config";

export function getWebsites() {
	const sheet =
		SpreadsheetApp.getActiveSpreadsheet().getSheetByName("websites");
	if (sheet === null) {
		return [];
	}

	const rows = sheet.getDataRange().getValues().slice(1);
	const websites = rows.flatMap((row) => {
		try {
			const url = row[0];
			const words = row[1].split("m");
			const website = new Website(url, words);
			return [website];
		} catch (e) {
			Logger.log(e);
			return [];
		}
	});

	return websites;
}

export function initConfig() {
	const sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName("config");
	if (sheet === null) {
		return null;
	}
	const values = sheet.getDataRange().getValues();

	const config: Config = {
		discordWebhookUrl: values[0][1],
		start: values[1][1],
		end: values[2][1],
	};
	return config;
}
