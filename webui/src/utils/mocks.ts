export const useFetchWithPaginationMock = (options = {}) => ({
  error: null,
  isEmpty: false,
  isLoadingMore: false,
  isReachingEnd: true,
  loadMore: vi.fn,
  pageCount: 1,
  pageSWRs: [],
  pages: null,
  ...options,
})
