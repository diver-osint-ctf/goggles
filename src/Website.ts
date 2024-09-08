type Result = {
	ok?: boolean;
	status?: number;
	hasWord?: boolean;
	date?: Date;
};
class Website {
	public result: Result;
	constructor(
		private url: string,
		private checkWords: string[],
	) {
		this.result = {};
	}

	public healthCheck() {
		const response = UrlFetchApp.fetch(this.url);

		const statusCode = response.getResponseCode();
		this.result.status = statusCode;
		const validStatusCode = statusCode < 200 || 300 <= statusCode;

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
