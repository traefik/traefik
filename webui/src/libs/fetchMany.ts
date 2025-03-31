/* eslint-disable @typescript-eslint/no-explicit-any */
import { Key } from 'swr'

const { VITE_APP_BASE_API_URL } = import.meta.env

export default async function <JSON = any>(key: Key): Promise<JSON[] | undefined> {
  const [baseUrl, params, init] = key as Array<string | string[] | RequestInit>

  if (!params || !Array.isArray(params)) return

  const requests = params.map((param) => {
    const apiUrl = `${window.APIUrl || VITE_APP_BASE_API_URL || ''}${baseUrl}${param}`
    return fetch(apiUrl, init as RequestInit).then((res) => res.json())
  })

  return await Promise.all(requests)
}
