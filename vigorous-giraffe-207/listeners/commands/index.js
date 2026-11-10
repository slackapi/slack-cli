import { sampleCommandCallback } from './sample-command.js';

export const register = (app) => {
  app.command('/sample-command', sampleCommandCallback);
};
