import assert from 'node:assert';
import { beforeEach, describe, it, mock } from 'node:test';
import { sampleShortcutCallback } from '../../listeners/shortcuts/sample-shortcut.js';

describe('shortcuts', () => {
  let fakeShortcut;
  let fakeAck;
  let fakeClient;
  let fakeLogger;

  beforeEach(() => {
    fakeShortcut = {
      trigger_id: 't1234',
    };
    fakeAck = mock.fn();
    fakeClient = {
      views: {
        open: mock.fn(),
      },
    };
    fakeLogger = {
      error: mock.fn(),
    };
  });

  it('should acknowledge and open modal view', async () => {
    await sampleShortcutCallback({
      shortcut: fakeShortcut,
      ack: fakeAck,
      client: fakeClient,
      logger: fakeLogger,
    });

    assert.strictEqual(fakeAck.mock.callCount(), 1);
    assert.strictEqual(fakeClient.views.open.mock.callCount(), 1);

    const callArgs = fakeClient.views.open.mock.calls[0].arguments[0];
    assert.equal(callArgs.trigger_id, fakeShortcut.trigger_id);
    assert(callArgs.view);
  });

  it('should log error when ack throws exception', async () => {
    const testError = new Error('test exception');
    fakeAck = mock.fn(() => {
      throw testError;
    });

    await sampleShortcutCallback({
      shortcut: fakeShortcut,
      ack: fakeAck,
      client: fakeClient,
      logger: fakeLogger,
    });

    assert.strictEqual(fakeClient.views.open.mock.callCount(), 0);
    assert.strictEqual(fakeAck.mock.callCount(), 1);
    assert.deepEqual(fakeLogger.error.mock.calls[0].arguments, [testError]);
  });
});
