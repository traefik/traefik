import { get } from 'dot-prop'

export default function PaginationMixin (opts = {}) {
  const { pollingIntervalTime, rowsPerPage = 10 } = opts
  let listLength = 0
  let currentPage = 1
  let currentLimit = rowsPerPage

  return {
    methods: {
      fetchWithInterval () {
        this.initFetch({ limit: listLength })
        this.pollingInterval = setInterval(
          () => {
            this.fetchMore({
              limit: Math.ceil(listLength / rowsPerPage) * rowsPerPage, // round up to multiple of rowsPerPage
              refresh: true
            })
          },
          pollingIntervalTime
        )
      },
      fetchMore ({ page = 1, limit = rowsPerPage, refresh, ...params } = {}) {
        if (page === currentPage && limit === currentLimit && !refresh) {
          return Promise.resolve()
        }

        currentPage = page
        currentLimit = limit || rowsPerPage

        const fetchMethod = get(this, opts.fetchMethod)

        return fetchMethod({
          ...params,
          page,
          limit: limit || rowsPerPage
        }).then(res => {
          listLength = page > 1
            ? listLength += res.data.length
            : res.data.length
        })
      },
      initFetch (params) {
        const scrollerRef = get(this.$refs, opts.scrollerRef)

        if (scrollerRef) {
          scrollerRef.stop()
          scrollerRef.reset()
        }

        return this.fetchMore({
          page: 1,
          refresh: true,
          ...params
        }).then(() => {
          if (scrollerRef) {
            scrollerRef.resume()
            scrollerRef.poll()
          }
        })
      }
    },
    mounted () {
      if (pollingIntervalTime) {
        this.fetchWithInterval()
      } else {
        this.fetchMore()
      }
    },
    beforeDestroy () {
      clearInterval(this.pollingInterval)
    }
  }
}
