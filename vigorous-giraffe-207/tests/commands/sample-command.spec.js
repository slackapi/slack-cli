import assert from 'node:assert';
import { beforeEach, describe, it, mock } from 'node:test';
import { sampleCommandCallback } from '../../listeners/commands/sample-command.js';

describe('commands', () => {
  let fakeAck;
  let fakeRespond;
  let fakeLogger;

  beforeEach(() => {
    fakeAck = mock.fn();
    fakeRespond = mock.fn();
    fakeLogger = {
      error: mock.fn(),
    };
  });

  it('should acknowledge and respond to command', async () => {
    await sampleCommandCallback({
      ack: fakeAck,
      respond: fakeRespond,
      logger: fakeLogger,
    });

    assert.strictEqual(fakeAck.mock.callCount(), 1);
    assert.strictEqual(fakeRespond.mock.callCount(), 1);

    const callArgs = fakeRespond.mock.calls[0].arguments[0];
    assert(callArgs.includes('Responding to the sample command!'));
  });

  it('should log error when ack throws exception', async () => {
    const testError = new Error('test exception');
    fakeAck = mock.fn(() => {
      throw testError;
    });

    await sampleCommandCallback({
      ack: fakeAck,
      respond: fakeRespond,
      logger: fakeLogger,
    });

    assert.strictEqual(fakeAck.mock.callCount(), 1);
    assert.deepEqual(fakeLogger.error.mock.calls[0].arguments, [testError]);
  });
});
