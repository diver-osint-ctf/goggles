type Result = {
	ok?: boolean;
	status?: number;
	hasWord?: boolean;
	date?: Date;
};
class Website {
	public result: Result;
	constructor(
		public url: string,
		public checkWords: string[],
	) {
		this.result = {};
	}

	public healthCheck() {
		Logger.log(`Checking ${this.url}...`);
		const response = UrlFetchApp.fetch(this.url, {
			followRedirects: true,
			muteHttpExceptions: true,
		});

		const statusCode = response.getResponseCode();
		this.result.status = statusCode;
		const validStatusCode = 200 <= statusCode && statusCode < 300;

		const content = response.getContentText();
		const hasWord = this.checkWords.some((word) => {
			return content.includes(word);
		});
		this.result.hasWord = hasWord;
		this.result.date = new Date();
		this.result.ok = validStatusCode && hasWord;
	}
}

export default Website;
