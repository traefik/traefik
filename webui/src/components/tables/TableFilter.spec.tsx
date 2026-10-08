import { fireEvent, screen, waitFor } from '@testing-library/react'
import { useSearchParams } from 'react-router-dom'

import { TableFilter } from './TableFilter'

import { renderWithProviders } from 'utils/test'

// Renders the query params TableFilter writes to the URL, which is what table pages
// actually read to fetch/filter data (see searchParamsToState usages in pages/*).
const SearchParamsProbe = () => {
  const [searchParams] = useSearchParams()
  return <div data-testid="params-probe">{searchParams.toString()}</div>
}

const renderTableFilter = (route: string, props: { hideStatusFilter?: boolean } = {}) =>
  renderWithProviders(
    <>
      <TableFilter {...props} />
      <SearchParamsProbe />
    </>,
    { route },
  )

const readStored = (pathname: string) => {
  const raw = localStorage.getItem(`traefik-table-filters-${pathname}`)
  return raw ? JSON.parse(raw) : null
}

const readProbe = () => Object.fromEntries(new URLSearchParams(screen.getByTestId('params-probe').textContent ?? ''))

describe('<TableFilter />', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('should start with an empty search and no status selected when nothing is stored', () => {
    renderTableFilter('/http/routers')

    expect(screen.getByTestId('table-search-input')).toHaveValue('')
    expect(readProbe()).toEqual({})
  })

  it('should persist search and status together in localStorage, scoped by pathname', async () => {
    renderTableFilter('/http/routers')

    fireEvent.change(screen.getByTestId('table-search-input'), { target: { value: 'traefik' } })

    await waitFor(() => expect(readStored('/http/routers')?.search).toBe('traefik'), { timeout: 2000 })

    fireEvent.click(screen.getByRole('button', { name: 'Warnings' }))

    await waitFor(() => {
      expect(readStored('/http/routers')).toEqual({ search: 'traefik', status: 'warning' })
    })
  })

  it('should restore both search and status on remount', () => {
    localStorage.setItem('traefik-table-filters-/http/routers', JSON.stringify({ search: 'traefik', status: 'warning' }))

    renderTableFilter('/http/routers')

    expect(screen.getByTestId('table-search-input')).toHaveValue('traefik')
    expect(readProbe()).toEqual({ search: 'traefik', status: 'warning' })
  })

  it('should not restore filters stored for a different page', () => {
    localStorage.setItem('traefik-table-filters-/http/services', JSON.stringify({ search: 'traefik', status: 'warning' }))

    renderTableFilter('/http/routers')

    expect(screen.getByTestId('table-search-input')).toHaveValue('')
    expect(readProbe()).toEqual({})
  })

  it('should let a URL param take precedence over its stored counterpart while still restoring the other field', () => {
    localStorage.setItem('traefik-table-filters-/http/routers', JSON.stringify({ search: 'from-storage', status: 'warning' }))

    renderTableFilter('/http/routers?status=disabled')

    expect(screen.getByTestId('table-search-input')).toHaveValue('from-storage')
    expect(readProbe()).toEqual({ search: 'from-storage', status: 'disabled' })
  })

  it('should drop only the cleared field, keeping the other one stored', async () => {
    localStorage.setItem('traefik-table-filters-/http/routers', JSON.stringify({ search: 'traefik', status: 'warning' }))

    renderTableFilter('/http/routers')

    fireEvent.change(screen.getByTestId('table-search-input'), { target: { value: '' } })

    await waitFor(() => {
      expect(readStored('/http/routers')).toEqual({ status: 'warning' })
    })
  })

  it('should remove the stored entry once both search and status are cleared', async () => {
    localStorage.setItem('traefik-table-filters-/http/routers', JSON.stringify({ search: 'traefik', status: 'warning' }))

    renderTableFilter('/http/routers')

    fireEvent.click(screen.getByRole('button', { name: 'All status' }))
    fireEvent.change(screen.getByTestId('table-search-input'), { target: { value: '' } })

    await waitFor(() => {
      expect(localStorage.getItem('traefik-table-filters-/http/routers')).toBeNull()
    })
  })
})
