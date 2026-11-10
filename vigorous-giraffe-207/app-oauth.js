import { App, FileInstallationStore, LogLevel } from '@slack/bolt';
import { config } from 'dotenv';
import { registerListeners } from './listeners/index.js';

config();

const app = new App({
  logLevel: LogLevel.DEBUG,
  signingSecret: process.env.SLACK_SIGNING_SECRET,
  clientId: process.env.SLACK_CLIENT_ID,
  clientSecret: process.env.SLACK_CLIENT_SECRET,
  stateSecret: 'my-state-secret',
  scopes: ['channels:history', 'chat:write', 'commands'],
  // FileInstallationStore is intended for development purposes only
  // Learn more: https://docs.slack.dev/tools/bolt-js/concepts/authenticating-oauth#installation-store
  installationStore: new FileInstallationStore(),
  installerOptions: {
    // If true, /slack/install redirects installers to the Slack Authorize URL
    // without rendering the web page with "Add to Slack" button
    directInstall: false,
  },
  clientOptions: {
    slackApiUrl: process.env.SLACK_API_URL,
  },
});

/** Register Listeners */
registerListeners(app);

/** Start Bolt App */
(async () => {
  try {
    await app.start(process.env.PORT || 3000);
    app.logger.info('⚡️ Bolt app is running! ⚡️');
  } catch (error) {
    app.logger.error('Unable to start App', error);
  }
})();
