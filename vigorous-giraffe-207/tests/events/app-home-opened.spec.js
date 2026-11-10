import assert from 'node:assert';
import { beforeEach, describe, it, mock } from 'node:test';
import { appHomeOpenedCallback } from '../../listeners/events/app-home-opened.js';

describe('events', () => {
  let fakeClient;
  let fakeEvent;
  let fakeLogger;

  beforeEach(() => {
    fakeClient = {
      views: {
        publish: mock.fn(),
      },
    };
    fakeEvent = {
      tab: 'home',
      user: 'U123',
    };
    fakeLogger = {
      error: mock.fn(),
    };
  });

  it('should publish home view when tab is home', async () => {
    await appHomeOpenedCallback({
      client: fakeClient,
      event: fakeEvent,
      logger: fakeLogger,
    });

    assert.strictEqual(fakeClient.views.publish.mock.callCount(), 1);

    const callArgs = fakeClient.views.publish.mock.calls[0].arguments[0];
    assert.equal(callArgs.user_id, fakeEvent.user);
    assert(callArgs.view);
  });

  it('should not publish when event tab is not home', async () => {
    fakeEvent.tab = 'about';

    await appHomeOpenedCallback({
      client: fakeClient,
      event: fakeEvent,
      logger: fakeLogger,
    });

    assert.strictEqual(fakeClient.views.publish.mock.callCount(), 0);
  });

  it('should log error when views publish throws exception', async () => {
    const testError = new Error('test exception');
    fakeClient.views.publish = mock.fn(() => {
      throw testError;
    });

    await appHomeOpenedCallback({
      client: fakeClient,
      event: fakeEvent,
      logger: fakeLogger,
    });

    assert.strictEqual(fakeClient.views.publish.mock.callCount(), 1);
    assert.deepEqual(fakeLogger.error.mock.calls[0].arguments, [testError]);
  });
});
