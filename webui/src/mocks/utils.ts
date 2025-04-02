import { chunk, cloneDeep, orderBy } from 'lodash'
import { http, HttpResponse } from 'msw'

const waitAsync = (noDelay = false) => {
  if (noDelay) return Promise.resolve()
  let delay = Math.random() + 0.5
  if (delay > 1) delay = 1
  return new Promise((res) => setTimeout(res, delay * 1000))
}

interface DataItem {
  name: string
  status?: string
}

export const listHandlers = (
  route: string,
  data: DataItem[] | Record<string, unknown> | null = null,
  noDelay: boolean = false,
  skipPagination = false,
) => [
  http.get(route, async ({ request }) => {
    await waitAsync(noDelay)
    const url = new URL(request.url)
    const direction = (url.searchParams.get('direction') as 'asc' | 'desc' | null) || 'asc'
    const search = url.searchParams.get('search')
    const sortBy = url.searchParams.get('sortBy') || 'name'
    const status = url.searchParams.get('status')
    let results = cloneDeep(data)
    if (Array.isArray(results)) {
      if (search) results = results.filter((x) => x.name.toLowerCase().includes(search.toLowerCase()))
      if (status) results = results.filter((x) => x.status === status)
      if (!results.length) return HttpResponse.json([], { headers: { 'X-Next-Page': '1' }, status: 200 })

      if (sortBy) results = orderBy(results as DataItem[], [sortBy], [direction || 'asc'])
      const page = +(url.searchParams.get('page') || 1)
      const pageSize = +(url.searchParams.get('per_page') || 10)
      const chunks = skipPagination ? [results] : chunk(results, pageSize)
      const totalPages = chunks.length
      const nextPage = page + 1 <= totalPages ? page + 1 : 1 // 1 means "no more pages".
      return HttpResponse.json(chunks[page - 1], { headers: { 'X-Next-Page': nextPage.toString() }, status: 200 })
    }
    return HttpResponse.json(results, { status: 200 })
  }),
  http.get(`${route}/:name`, async ({ params }) => {
    await waitAsync(noDelay)

    if (!Array.isArray(data)) {
      return HttpResponse.json({}, { status: 501 })
    }

    const { name } = params
    const res = data.find((x) => x.name === name)
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
