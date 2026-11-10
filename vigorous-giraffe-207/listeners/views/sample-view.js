const sampleViewCallback = async ({ ack, view, body, client, logger }) => {
  try {
    await ack();

    const formValues = view.state.values;
    const sampleInputValue = formValues.input_block_id.sample_input_id.value;
    const sampleConvoDropdown = formValues.select_channel_block_id.sample_dropdown_id;

    await client.chat.postMessage({
      channel: sampleConvoDropdown.selected_conversation,
      text: `<@${body.user.id}> submitted the following :sparkles: hopes and dreams :sparkles:: \n\n ${sampleInputValue}`,
    });
  } catch (error) {
    logger.error(error);
  }
};

export { sampleViewCallback };
