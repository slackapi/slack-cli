import { sampleViewCallback } from './sample-view.js';

export const register = (app) => {
  app.view('sample_view_id', sampleViewCallback);
};
