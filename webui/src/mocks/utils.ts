/* eslint-disable @typescript-eslint/no-explicit-any */

import { chunk, cloneDeep } from 'lodash'
import { http, HttpResponse } from 'msw'

const waitAsync = (noDelay = false) => {
  if (noDelay) return Promise.resolve()
  let delay = Math.random() + 0.5
  if (delay > 1) delay = 1
  return new Promise((res) => setTimeout(res, delay * 1000))
}

export const listHandlers = (route: string, data: any = null, noDelay: boolean = false, skipPagination = false) => [
  http.get(route, async ({ request }) => {
    await waitAsync(noDelay)
    const url = new URL(request.url)
    const search = url.searchParams.get('search')
    const status = url.searchParams.get('status')
    const results =
      Array.isArray(data) && (search || status)
        ? data
            .filter((x: any) => (search ? x.name.toLowerCase().includes(search.toLowerCase()) : true))
            .filter((x: any) => (status ? x.status === status : true))
        : cloneDeep(data)
    if (Array.isArray(results)) {
      if (!results.length) return HttpResponse.json([], { headers: { 'X-Next-Page': '1' }, status: 200 })
      const page = +(url.searchParams.get('page') || 1)
      const pageSize = +(url.searchParams.get('per_page') || 10)
      const chunks = skipPagination ? [results] : chunk(results, pageSize)
      const totalPages = chunks.length
      const nextPage = page + 1 <= totalPages ? page + 1 : 1 // 1 means "no more pages".
      return HttpResponse.json(chunks[page - 1], { headers: { 'X-Next-Page': nextPage.toString() }, status: 200 })
    } else {
      return HttpResponse.json(results, { status: 200 })
    }
  }),
  http.get(`${route}/:name`, async ({ params }) => {
    await waitAsync(noDelay)

    if (!Array.isArray(data)) {
      return HttpResponse.json({}, { status: 501 })
    }

    const { name } = params
    const res = data.find((x: any) => x.name === name)
    if (!res) {
      const parts = route.split('/')
      const lastPart = parts[parts.length - 1]
      return HttpResponse.json(
        {
          message: `${lastPart.substring(0, lastPart.length - 1)} not found: ${name}`,
        },
        { status: 404 },
      )
    }
    return HttpResponse.json(res, { status: 200 })
  }),
]
