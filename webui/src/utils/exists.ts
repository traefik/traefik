// eslint-disable-next-line @typescript-eslint/no-explicit-any
const exists = (param: any): boolean => {
  return typeof param !== 'undefined'
}

export default exists
