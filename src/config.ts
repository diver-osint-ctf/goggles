export type Config = {
	discordWebhookUrl: string;
	start: Date;
	end: Date;
  mentions: string[];
};

export const env = {
	SHEET_URL: PropertiesService.getScriptProperties().getProperty("SHEET_URL"),
};

export const REPORT_TEMPLATE = `# Failed Websites Report
{FAILURES}

Reported by [Goggles](https://github.com/diver-osint-ctf/goggles)`;

export const WEBSITE_REPORT_TEMPLATE = `- url: {URL}
{CHECK_WORDS}
  - factor
    - HTTP Response Code: {STATUS}
    - Has Check Words: {HAS_WORD}`;

export const CHECK_WORDS_TEMPLATE = `  - check words:
{WORDS}`;

export const WORDS_TEMPLATE = "    - {WORD}";

export const SUCCESS_COLOR = 3066993; // #2ECC71
export const FAILURE_COLOR = 15158332; // #E74C3C
