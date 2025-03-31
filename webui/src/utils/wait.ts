// eslint-disable-next-line no-undef
export const wait = (ms: number): Promise<void> => new Promise((resolve) => setTimeout(resolve, ms))
