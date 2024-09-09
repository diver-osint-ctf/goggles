import {
	CHECK_WORDS_TEMPLATE,
	env,
	FAILURE_COLOR,
	REPORT_TEMPLATE,
	SUCCESS_COLOR,
	WEBSITE_REPORT_TEMPLATE,
	WORDS_TEMPLATE,
	type Config,
} from "./config";
import type Website from "./Website";

class Discord {
	constructor(private config: Config) {}

	private makeEmbedsContent(failures: Website[], websiteCount: number) {
		if (failures.length === 0) {
			return [
				{
					title: `Nice! ${websiteCount} All websites are healthy :tada:`,
					description: `If you want to see list of websites, click the [link](${env.SHEET_URL}).`,
					color: SUCCESS_COLOR,
				},
			];
		}
		return [
			{
				title: `${failures.length} Website is unhealthy :tired_face:`,
				description: `If you want to see list of websites, click the [link](${env.SHEET_URL}).`,
				color: FAILURE_COLOR,
				author: {
					name: "Goggles",
					url: "https://github.com/diver-osint-ctf/goggles",
				},
			},
		];
	}

	sendSummary(websites: Website[]) {
		const failures = websites.filter((website) => !website.result.ok);
		const embedsContent = this.makeEmbedsContent(failures, websites.length);
		const content = {
			embeds: embedsContent,
			tts: false,
		};
		const options: GoogleAppsScript.URL_Fetch.URLFetchRequestOptions = {
			method: "post",
			headers: {
				"Content-Type": "application/json",
			},
			payload: JSON.stringify(content),
		};
		UrlFetchApp.fetch(this.config.discordWebhookUrl, options);
		return failures;
	}

	private makeReportContent(websites: Website[]) {
		const failures = websites
			.filter((website) => !website.result.ok)
			.map((website) => {
				const words = website.checkWords
					.map((word) => WORDS_TEMPLATE.replace("{WORD}", word))
					.join("\n");
				const checkWords = CHECK_WORDS_TEMPLATE.replace("{WORDS}", words);
				return WEBSITE_REPORT_TEMPLATE.replace("{URL}", website.url)
					.replace("{CHECK_WORDS}", checkWords)
					.replace("{STATUS}", (website.result.status ?? "unknown").toString())
					.replace(
						"{HAS_WORD}",
						(website.result.hasWord ?? "unknown").toString(),
					);
			})
			.join("\n");

		return REPORT_TEMPLATE.replace("{FAILURES}", failures);
	}

	async sendReport(websites: Website[]) {
		const fileName = "report.md";
		const fileContent = this.makeReportContent(websites);
		const formData = {
			file: Utilities.newBlob(fileContent, "text/plain", fileName),
			payload_json: JSON.stringify({
				content: this.config.mentions
					.map((mention) => `<@${mention}>`)
					.join(" "),
			}),
		};

		const options: GoogleAppsScript.URL_Fetch.URLFetchRequestOptions = {
			method: "post",
			payload: formData,
			muteHttpExceptions: true,
		};

		Logger.log("Sending report to Discord...");
		const response = UrlFetchApp.fetch(this.config.discordWebhookUrl, options);
		if (response.getResponseCode() !== 200) {
			Logger.log(response.getContentText());
		}
	}

	sendError(text: string) {
		const content = {
			content: `<@${this.config.mentions?.[0]}> [ERROR] ${text}`,
		};
		const options: GoogleAppsScript.URL_Fetch.URLFetchRequestOptions = {
			method: "post",
			headers: {
				"Content-Type": "application/json",
			},
			payload: JSON.stringify(content),
		};
		UrlFetchApp.fetch(this.config.discordWebhookUrl, options);
	}
}

export default Discord;
