import { BASE_PATH } from './utils'

export default async function (input: RequestInfo, init?: RequestInit): Promise<JSON> {
  const res = await fetch(`${BASE_PATH}${input}`, init)
  if (!res.ok) throw new Error(res.statusText)
  return await res.json()
}

export const fetchPage = async function (
  input: RequestInfo,
  init?: RequestInit,
): Promise<Response & { data: unknown[]; nextPage: number }> {
  const res = await fetch(`${BASE_PATH}${input}`, init)

  if (!res.ok) throw new Error(res.statusText)

  return res.json().then((data) => {
    return {
      ...res,
      data,
      nextPage: parseInt(res.headers.get('X-Next-Page') || '1'),
    }
  })
}
