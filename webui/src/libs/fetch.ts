export default async function (input: RequestInfo, init?: RequestInit): Promise<JSON> {
  const { VITE_APP_BASE_API_URL } = import.meta.env
  const res = await fetch(`${window.APIUrl || VITE_APP_BASE_API_URL || ''}${input}`, init)
  if (!res.ok) throw new Error(res.statusText)
  return await res.json()
}

export const fetchPage = async function (
  input: RequestInfo,
  init?: RequestInit,
): Promise<Response & { data: unknown[]; nextPage: number }> {
  const { VITE_APP_BASE_API_URL } = import.meta.env
  const res = await fetch(`${window.APIUrl || VITE_APP_BASE_API_URL || ''}${input}`, init)

  if (!res.ok) throw new Error(res.statusText)

  return res.json().then((data) => {
    return {
      ...res,
      data,
      nextPage: parseInt(res.headers.get('X-Next-Page') || '1'),
    }
  })
}
