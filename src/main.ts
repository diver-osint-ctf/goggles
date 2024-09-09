import Discord from "./Discord";
import { getWebsites, initConfig } from "./sheets";

export async function healthCheck() {
	Logger.log("healthCheck is running...");
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
