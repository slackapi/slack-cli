import { sampleShortcutCallback } from './sample-shortcut.js';

export const register = (app) => {
  app.shortcut('sample_shortcut_id', sampleShortcutCallback);
};
