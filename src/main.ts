import Discord from "./Discord";
import { getWebsites, initConfig } from "./sheets";

export async function healthCheck() {
	const config = initConfig();
	if (config === null) {
		Logger.log("config is null");
		return;
	}

	const websites = getWebsites();
	if (websites.length === 0) {
		return;
	}

	for (const website of websites) {
		website.healthCheck();
	}

	const discord = new Discord(config);
	const failures = discord.sendSummary(websites);
	if (failures.length > 0) {
		await discord.sendReport(websites);
	}
	return "[INFO] done.";
}

export async function regularExecution() {
	const config = initConfig();
	if (config === null) {
		Logger.log("config is null");
		return;
	}
	try {
		const now = Date.now();
		const startDate = Date.parse(config.start);
		const endDate = Date.parse(config.end);
		if (Number.isNaN(startDate) || Number.isNaN(endDate)) {
			throw new Error(
				`Invalid date format: start=${config.start}, end=${config.end}`,
			);
		}
		if (startDate <= now && now <= endDate) {
			await healthCheck();
		}
	} catch (e) {
		Logger.log(e);
		if (e instanceof Error) {
			const discord = new Discord(config);
			discord.sendError(e.toString());
		}
	}
}
