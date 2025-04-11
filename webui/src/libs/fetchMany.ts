import { Key } from 'swr'

import { BASE_PATH } from './utils'

export default async function <JSON>(key: Key): Promise<JSON[] | undefined> {
  const [baseUrl, params, init] = key as Array<string | string[] | RequestInit>

  if (!params || !Array.isArray(params)) return

  const requests = params.map((param) => {
    const apiUrl = `${BASE_PATH}${baseUrl}${param}`
    return fetch(apiUrl, init as RequestInit).then((res) => res.json())
  })

  return await Promise.all(requests)
}
