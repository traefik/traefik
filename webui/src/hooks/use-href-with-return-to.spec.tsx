import { renderHook } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'

import { useGetUrlWithReturnTo, useHrefWithReturnTo, useRouterReturnTo } from './use-href-with-return-to'

describe('useGetUrlWithReturnTo', () => {
  const createWrapper = (initialPath = '/') => {
    return ({ children }) => <MemoryRouter initialEntries={[initialPath]}>{children}</MemoryRouter>
  }

  it('should append current path as returnTo query param', () => {
    const { result } = renderHook(() => useGetUrlWithReturnTo('/target'), {
      wrapper: createWrapper('/current/path'),
    })

    expect(result.current).toBe('/target?returnTo=%2Fcurrent%2Fpath')
  })

  it('should append current path with search params as returnTo', () => {
    const { result } = renderHook(() => useGetUrlWithReturnTo('/target'), {
      wrapper: createWrapper('/current/path?foo=bar'),
    })

    expect(result.current).toBe('/target?returnTo=%2Fcurrent%2Fpath%3Ffoo%3Dbar')
  })

  it('should use initialReturnTo when provided', () => {
    const { result } = renderHook(() => useGetUrlWithReturnTo('/target', '/custom/return'), {
      wrapper: createWrapper('/current/path'),
    })

    expect(result.current).toBe('/target?returnTo=%2Fcustom%2Freturn')
  })

  it('should return the href as-is when href is empty string', () => {
    const { result } = renderHook(() => useGetUrlWithReturnTo(''), {
      wrapper: createWrapper('/current/path'),
    })

    expect(result.current).toBe('')
  })

  it('should handle href with existing query params', () => {
    const { result } = renderHook(() => useGetUrlWithReturnTo('/target?existing=param'), {
      wrapper: createWrapper('/current/path'),
    })

    expect(result.current).toBe('/target?existing=param&returnTo=%2Fcurrent%2Fpath')
  })
})

describe('useHrefWithReturnTo', () => {
  const createWrapper = (initialPath = '/') => {
    return ({ children }) => <MemoryRouter initialEntries={[initialPath]}>{children}</MemoryRouter>
  }

  it('should return resolved href with returnTo param containing current path', () => {
    const { result } = renderHook(() => useHrefWithReturnTo('/target'), {
      wrapper: createWrapper('/current'),
    })

    expect(result.current).toBe('/target?returnTo=%2Fcurrent')
  })

  it('should include current search params in returnTo', () => {
    const { result } = renderHook(() => useHrefWithReturnTo('/target'), {
      wrapper: createWrapper('/current?foo=bar&baz=qux'),
    })

    expect(result.current).toBe('/target?returnTo=%2Fcurrent%3Ffoo%3Dbar%26baz%3Dqux')
  })

  it('should use custom returnTo when provided instead of current path', () => {
    const { result } = renderHook(() => useHrefWithReturnTo('/target', '/custom/return'), {
      wrapper: createWrapper('/current'),
    })

    expect(result.current).toBe('/target?returnTo=%2Fcustom%2Freturn')
  })

  it('should handle absolute paths correctly', () => {
    const { result } = renderHook(() => useHrefWithReturnTo('/http/routers'), {
      wrapper: createWrapper('/tcp/services'),
    })

    expect(result.current).toBe('/http/routers?returnTo=%2Ftcp%2Fservices')
  })

  it('should preserve existing query params in target href', () => {
    const { result } = renderHook(() => useHrefWithReturnTo('/target?existing=param'), {
      wrapper: createWrapper('/current'),
    })

    expect(result.current).toBe('/target?existing=param&returnTo=%2Fcurrent')
  })

  it('should return root path when href is empty', () => {
    const { result } = renderHook(() => useHrefWithReturnTo(''), {
      wrapper: createWrapper('/current'),
    })

    // useHref converts empty string to root path
    expect(result.current).toBe('/')
  })

  it('should handle complex nested paths in returnTo', () => {
    const { result } = renderHook(() => useHrefWithReturnTo('/target'), {
      wrapper: createWrapper('/http/routers/my-router-123'),
    })

    expect(result.current).toBe('/target?returnTo=%2Fhttp%2Frouters%2Fmy-router-123')
  })
})

describe('useRouterReturnTo', () => {
  const createWrapper = (initialPath = '/') => {
    return ({ children }) => <MemoryRouter initialEntries={[initialPath]}>{children}</MemoryRouter>
  }

  it('should return null when no returnTo query param exists', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current'),
    })

    expect(result.current.returnTo).toBeNull()
    expect(result.current.returnToLabel).toBeNull()
  })

  it('should extract returnTo from query params', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current?returnTo=/http/routers'),
    })

    expect(result.current.returnTo).toBe('/http/routers')
  })

  it('should generate correct label for HTTP routers (plural)', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current?returnTo=/http/routers'),
    })

    expect(result.current.returnToLabel).toBe('HTTP routers')
  })

  it('should generate correct label for HTTP router (singular)', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current?returnTo=/http/routers/router-1'),
    })

    expect(result.current.returnToLabel).toBe('HTTP router')
  })

  it('should generate fallback label for unknown routes (plural)', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current?returnTo=/custom/resources'),
    })

    expect(result.current.returnToLabel).toBe('Custom resources')
  })

  it('should handle malformed returnTo paths gracefully', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current?returnTo=/'),
    })

    expect(result.current.returnTo).toBe('/')
    expect(result.current.returnToLabel).toBe('Back')
  })

  it('should handle returnTo with query params', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current?returnTo=/http/routers?filter=test'),
    })

    expect(result.current.returnTo).toContain('/http/routers')
    expect(result.current.returnToLabel).toBe('HTTP routers')
  })

  it('should strip query params from path when generating label', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current?returnTo=/http/routers?filter=test&status=active'),
    })

    expect(result.current.returnToLabel).toBe('HTTP routers')
    expect(result.current.returnToLabel).not.toContain('filter')
    expect(result.current.returnToLabel).not.toContain('status')
  })

  it('should strip query params from subpath when generating label', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current?returnTo=/tcp/services?page=2'),
    })

    expect(result.current.returnToLabel).toBe('TCP services')
  })

  it('should handle query params with multiple question marks gracefully', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current?returnTo=/http/routers?filter=test?extra=param'),
    })

    // Should handle edge case with multiple question marks (invalid URL but should not crash)
    expect(result.current.returnToLabel).toBe('HTTP routers')
  })

  it('should handle path with query params but no subpath', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current?returnTo=/http?foo=bar'),
    })

    expect(result.current.returnToLabel).toBe('Http')
  })

  it('should handle empty query string (path ending with ?)', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current?returnTo=/tcp/middlewares?'),
    })

    expect(result.current.returnToLabel).toBe('TCP middlewares')
  })

  it('should handle complex query strings with special characters', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current?returnTo=/http/services?filter=%40test%23special'),
    })

    expect(result.current.returnToLabel).toBe('HTTP services')
  })

  it('should capitalize first letter of label override', () => {
    const { result } = renderHook(() => useRouterReturnTo(), {
      wrapper: createWrapper('/current?returnTo=/resource/routers/router-1'),
    })

    // Verify the label starts with uppercase
    expect(result.current.returnToLabel?.charAt(0)).toBe('R')
  })
})
