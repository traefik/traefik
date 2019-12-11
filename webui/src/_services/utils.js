export const getTotal = (headers, params) => {
  const nextPage = parseInt(headers['x-next-page'], 10) || 1
  const hasNextPage = nextPage > 1

  return hasNextPage
    ? (params.page + 1) * params.limit
    : params.page * params.limit
}
