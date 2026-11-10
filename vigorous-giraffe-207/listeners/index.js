import * as actions from './actions/index.js';
import * as commands from './commands/index.js';
import * as events from './events/index.js';
import * as messages from './messages/index.js';
import * as shortcuts from './shortcuts/index.js';
import * as views from './views/index.js';

export const registerListeners = (app) => {
  actions.register(app);
  commands.register(app);
  events.register(app);
  messages.register(app);
  shortcuts.register(app);
  views.register(app);
};
