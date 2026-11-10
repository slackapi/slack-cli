const sampleCommandCallback = async ({ ack, respond, logger }) => {
  try {
    await ack();
    await respond('Responding to the sample command!');
  } catch (error) {
    logger.error(error);
  }
};

export { sampleCommandCallback };
