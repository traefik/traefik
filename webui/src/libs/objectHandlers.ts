/* eslint-disable @typescript-eslint/no-explicit-any */
import exists from 'utils/exists'

type ObjectWithMessage = {
  message?: string
}

export const getValidData = <T extends ObjectWithMessage>(data?: T[]): T[] =>
  data ? data.filter((item) => !exists(item.message)) : []
export const getErrorData = <T extends ObjectWithMessage>(data?: T[]): T[] =>
  data ? data.filter((item) => exists(item.message)) : []
