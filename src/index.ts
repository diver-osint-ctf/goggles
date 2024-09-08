import { healthCheck } from "./main";

declare const global: {
	[x: string]: unknown;
};

global.healthCheck = healthCheck;
