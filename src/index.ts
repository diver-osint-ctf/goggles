import { healthCheck, regularExecution } from "./main";

declare const global: {
	[x: string]: unknown;
};

global.healthCheck = healthCheck;
global.regularExecution = regularExecution;
