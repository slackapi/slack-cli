import { appHomeOpenedCallback } from './app-home-opened.js';

export const register = (app) => {
  app.event('app_home_opened', appHomeOpenedCallback);
};
