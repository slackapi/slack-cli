import assert from 'node:assert';
import { beforeEach, describe, it, mock } from 'node:test';
import { sampleViewCallback } from '../../listeners/views/sample-view.js';

describe('views', () => {
  let fakeAck;
  let fakeView;
  let fakeBody;
  let fakeClient;
  let fakeLogger;

  beforeEach(() => {
    fakeAck = mock.fn();
    fakeView = {
      state: {
        values: {
          input_block_id: {
            sample_input_id: {
              value: 'test value',
            },
          },
          select_channel_block_id: {
            sample_dropdown_id: {
              selected_conversation: 'C1234',
            },
          },
        },
      },
    };
    fakeBody = {
      user: {
        id: 'U1234',
      },
    };
    fakeClient = {
      chat: {
        postMessage: mock.fn(),
      },
    };
    fakeLogger = {
      error: mock.fn(),
    };
  });

  it('should acknowledge and post message with form values', async () => {
    await sampleViewCallback({
      ack: fakeAck,
      view: fakeView,
      body: fakeBody,
      client: fakeClient,
      logger: fakeLogger,
    });

    assert.strictEqual(fakeAck.mock.callCount(), 1);
    assert.strictEqual(fakeClient.chat.postMessage.mock.callCount(), 1);

    const callArgs = fakeClient.chat.postMessage.mock.calls[0].arguments[0];
    assert.equal(
      callArgs.channel,
      fakeView.state.values.select_channel_block_id.sample_dropdown_id.selected_conversation,
    );
    assert(callArgs.text.includes(fakeView.state.values.input_block_id.sample_input_id.value));
    assert(callArgs.text.includes(fakeBody.user.id));
  });

  it('should log error when ack throws exception', async () => {
    const testError = new Error('test exception');
    fakeAck = mock.fn(() => {
      throw testError;
    });

    await sampleViewCallback({
      ack: fakeAck,
      view: fakeView,
      body: fakeBody,
      client: fakeClient,
      logger: fakeLogger,
    });

    assert.strictEqual(fakeAck.mock.callCount(), 1);
    assert.deepEqual(fakeLogger.error.mock.calls[0].arguments, [testError]);
  });
});
