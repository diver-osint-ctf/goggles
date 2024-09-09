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

	const text = `テストです\n\n[debug]\n- ${websites
		.map((website) => {
			website.healthCheck();
			return JSON.stringify(website.result);
		})
		.join("\n- ")}`;

  
  const discord = new Discord(config);
	discord.send(text);
  // await discord.sendFile();
  return "[INFO] done.";
}
