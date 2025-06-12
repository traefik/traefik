import { act, fireEvent, renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { SWRConfig } from 'swr'

import useFetchWithPagination from './use-fetch-with-pagination'

import { server } from 'mocks/server'
import { renderWithProviders } from 'utils/test'

const renderRow = (row) => (
  <li key={row.id} data-testid="listRow">
    {row.id}
  </li>
)

const wrapper = ({ children }) => (
  <SWRConfig
    value={{
      revalidateOnFocus: false,
      fetcher: fetch,
    }}
  >
    {children}
  </SWRConfig>
)

describe('useFetchWithPagination Hook', () => {
  it('should fetch 1st page per default', async () => {
    server.use(
      http.get('/api/http/routers', () => {
        return HttpResponse.json([{ id: 1 }], { status: 200 })
      }),
    )

    const { result } = renderHook(() => useFetchWithPagination('/http/routers', { renderRow }), {
      wrapper,
    })

    await waitFor(() => {
      expect(result.current.pages).not.toBeUndefined()
    })
  })

  it('should work as expected passing rowsPerPage property', async () => {
    let perPage

    server.use(
      http.get('/api/http/routers', ({ request }) => {
        const url = new URL(request.url)
        perPage = url.searchParams.get('per_page')
        return HttpResponse.json([{ id: 1 }], { status: 200 })
      }),
    )

    const { result } = renderHook(() => useFetchWithPagination('/http/routers', { renderRow, rowsPerPage: 3 }), {
      wrapper,
    })

    await waitFor(() => {
      expect(result.current.pages).not.toBeUndefined()
    })

    expect(perPage).toBe('3')
  })

  it('should work as expected requesting page 2', async () => {
    server.use(
      http.get('/api/http/routers', ({ request }) => {
        const url = new URL(request.url)
        const page = url.searchParams.get('page')
        if (page === '2') {
          return HttpResponse.json([{ id: 3 }], {
            headers: {
              'X-Next-Page': '1',
            },
            status: 200,
          })
        }
        return HttpResponse.json([{ id: 1 }, { id: 2 }], {
          headers: {
            'X-Next-Page': '2',
          },
          status: 200,
        })
      }),
    )

    const TestComponent = () => {
      const { pages, pageCount, loadMore, isLoadingMore } = useFetchWithPagination('/http/routers', {
        renderLoader: () => null,
        renderRow,
        rowsPerPage: 2,
      })

      return (
        <>
          <ul>{pages}</ul>
          {isLoadingMore ? <div data-testid="loading">Loading...</div> : <button onClick={loadMore}>Load More</button>}
          <div data-testid="pageCount">{pageCount}</div>
        </>
      )
    }

    const { queryAllByTestId, getByTestId, getByText } = renderWithProviders(<TestComponent />)

    await waitFor(() => {
      expect(() => {
        getByTestId('loading')
      }).toThrow('Unable to find an element by: [data-testid="loading"]')
    })

    act(() => {
      fireEvent.click(getByText(/Load More/))
    })

    await waitFor(() => {
      expect(() => {
        getByTestId('loading')
      }).toThrow('Unable to find an element by: [data-testid="loading"]')
    })

    expect(getByTestId('pageCount').innerHTML).toBe('2')

    const items = await queryAllByTestId('listRow')
    expect(items).toHaveLength(3)
  })

  it('should work as expected requesting an empty page', async () => {
    server.use(
      http.get('/api/http/routers', ({ request }) => {
        const url = new URL(request.url)
        const page = url.searchParams.get('page')
        if (page === '2') {
          return HttpResponse.json(
            // Response body should be { message: 'invalid request: page: 2, per_page: 4' }, resulting in a type error.
            // If I type the response body accordingly, allowing both an array and an object, MSW breaks, so I replaced
            // the object with an empty array, and that'd be enough for testing purpose.
            [],
            {
              headers: {
                'X-Next-Page': '1',
              },
              status: 200,
            },
          )
        }
        return HttpResponse.json([{ id: 1 }, { id: 2 }, { id: 3 }, { id: 4 }], {
          headers: {
            'X-Next-Page': '2',
          },
          status: 200,
        })
      }),
    )

    const TestComponent = () => {
      const { pages, pageCount, loadMore, isLoadingMore } = useFetchWithPagination('/http/routers', {
        renderLoader: () => null,
        renderRow,
        rowsPerPage: 4,
      })

      return (
        <>
          <ul>{pages}</ul>
          {isLoadingMore ? <div data-testid="loading">Loading...</div> : <button onClick={loadMore}>Load More</button>}
          <div data-testid="pageCount">{pageCount}</div>
        </>
      )
    }

    const { queryAllByTestId, getByTestId, getByText } = renderWithProviders(<TestComponent />)

    await waitFor(() => {
      expect(() => {
        getByTestId('loading')
      }).toThrow('Unable to find an element by: [data-testid="loading"]')
    })

    act(() => {
      fireEvent.click(getByText(/Load More/))
    })

    await waitFor(() => {
      expect(() => {
        getByTestId('loading')
      }).toThrow('Unable to find an element by: [data-testid="loading"]')
    })

    expect(getByTestId('pageCount').innerHTML).toBe('2')

    const items = await queryAllByTestId('listRow')
    expect(items).toHaveLength(4)
  })
})
