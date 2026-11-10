import assert from 'node:assert';
import { beforeEach, describe, it, mock } from 'node:test';
import { sampleActionCallback } from '../../listeners/actions/sample-action.js';

describe('actions', () => {
  let fakeAck;
  let fakeClient;
  let fakeBody;
  let fakeLogger;

  beforeEach(() => {
    fakeAck = mock.fn();
    fakeClient = {
      views: {
        update: mock.fn(),
      },
    };
    fakeBody = {
      view: {
        id: 'test_id',
        hash: '156772938.1827394',
      },
    };
    fakeLogger = {
      error: mock.fn(),
    };
  });

  it('should acknowledge and update view', async () => {
    await sampleActionCallback({
      ack: fakeAck,
      client: fakeClient,
      body: fakeBody,
      logger: fakeLogger,
    });

    assert.strictEqual(fakeAck.mock.callCount(), 1);
    assert.strictEqual(fakeClient.views.update.mock.callCount(), 1);

    const callArgs = fakeClient.views.update.mock.calls[0].arguments[0];
    assert.equal(callArgs.view_id, fakeBody.view.id);
    assert.equal(callArgs.hash, fakeBody.view.hash);
    assert(callArgs.view);
  });

  it('should log error when ack throws exception', async () => {
    const testError = new Error('test exception');
    fakeAck = mock.fn(() => {
      throw testError;
    });

    await sampleActionCallback({
      ack: fakeAck,
      client: fakeClient,
      body: fakeBody,
      logger: fakeLogger,
    });

    assert.strictEqual(fakeAck.mock.callCount(), 1);
    assert.deepEqual(fakeLogger.error.mock.calls[0].arguments, [testError]);
  });
});
