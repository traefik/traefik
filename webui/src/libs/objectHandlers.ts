type ObjectWithMessage = {
  message?: string
}

export const getValidData = <T extends ObjectWithMessage>(data?: T[]): T[] =>
  data ? data.filter((item) => !item.message) : []
export const getErrorData = <T extends ObjectWithMessage>(data?: T[]): T[] =>
  data ? data.filter((item) => !!item.message) : []
