import assert from 'node:assert';
import { beforeEach, describe, it, mock } from 'node:test';
import { sampleMessageCallback } from '../../listeners/messages/sample-message.js';

describe('messages', () => {
  let fakeContext;
  let fakeSay;
  let fakeLogger;

  beforeEach(() => {
    fakeContext = {
      matches: ['hello'],
    };
    fakeSay = mock.fn();
    fakeLogger = {
      error: mock.fn(),
    };
  });

  it('should call say with greeting response', async () => {
    await sampleMessageCallback({
      context: fakeContext,
      say: fakeSay,
      logger: fakeLogger,
    });

    assert.strictEqual(fakeSay.mock.callCount(), 1);
    const callArgs = fakeSay.mock.calls[0].arguments[0];
    assert(callArgs.includes(fakeContext.matches[0]));
    assert(callArgs.includes('hello, how are you?'));
  });

  it('should log error when say throws exception', async () => {
    const testError = new Error('test exception');
    fakeSay = mock.fn(() => {
      throw testError;
    });
    await sampleMessageCallback({
      context: fakeContext,
      say: fakeSay,
      logger: fakeLogger,
    });

    assert.strictEqual(fakeSay.mock.callCount(), 1);
    assert.deepEqual(fakeLogger.error.mock.calls[0].arguments, [testError]);
  });
});
