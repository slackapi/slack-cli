const sampleMessageCallback = async ({ context, say, logger }) => {
  try {
    const greeting = context.matches[0];
    await say(`${greeting}, how are you?`);
  } catch (error) {
    logger.error(error);
  }
};

export { sampleMessageCallback };
