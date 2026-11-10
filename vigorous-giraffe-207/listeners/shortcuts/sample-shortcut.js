const sampleShortcutCallback = async ({ shortcut, ack, client, logger }) => {
  try {
    const { trigger_id } = shortcut;

    await ack();
    await client.views.open({
      trigger_id,
      view: {
        type: 'modal',
        callback_id: 'sample_view_id',
        title: {
          type: 'plain_text',
          text: 'Sample modal title',
        },
        blocks: [
          {
            type: 'section',
            text: {
              type: 'mrkdwn',
              text: 'Click the button to update the modal',
            },
            accessory: {
              type: 'button',
              text: {
                type: 'plain_text',
                text: 'Update modal',
              },
              action_id: 'sample_action_id',
            },
          },
          {
            type: 'input',
            block_id: 'input_block_id',
            label: {
              type: 'plain_text',
              text: 'What are your hopes and dreams?',
            },
            element: {
              type: 'plain_text_input',
              action_id: 'sample_input_id',
              multiline: true,
            },
          },
          {
            block_id: 'select_channel_block_id',
            type: 'input',
            label: {
              type: 'plain_text',
              text: 'Select a channel to message the result to',
            },
            element: {
              type: 'conversations_select',
              action_id: 'sample_dropdown_id',
              response_url_enabled: true,
            },
          },
        ],
        submit: {
          type: 'plain_text',
          text: 'Submit',
        },
      },
    });
  } catch (error) {
    logger.error(error);
  }
};

export { sampleShortcutCallback };
