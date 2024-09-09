import type { Config } from "./config";

class Discord {
  constructor(private config: Config) {}

  send(text: string) {
    const content = {
      embeds: [
        {
          title: " Health Check by Goggles",
          description: text,
          color: 15158332,
          url: "https://github.com/diver-osint-ctf/goggles.git",
        },
      ],
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
  }

  async sendFile() {
    const fileContent =
      "This is a long string that will be sent as a file content.\n\n- aaa\n- aaa\n- aaa\n- aaa\n- aaa\n- aaa\n\naaaaaa";
    const fileName = "result.md";
    const formData = {
      file: Utilities.newBlob(fileContent, "text/plain", fileName)
    };

    const options: GoogleAppsScript.URL_Fetch.URLFetchRequestOptions = {
      method: "post",
      payload: formData,
      muteHttpExceptions: true,
    };

    Logger.log("Sending report to Discord...");
    const response = UrlFetchApp.fetch(this.config.discordWebhookUrl, options);

    Logger.log(response.getContentText());
  }
}

export default Discord;