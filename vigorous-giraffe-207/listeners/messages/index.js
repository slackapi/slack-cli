import { sampleMessageCallback } from './sample-message.js';

export const register = (app) => {
  app.message(/^(hi|hello|hey).*/, sampleMessageCallback);
};
