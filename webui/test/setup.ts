import '@testing-library/jest-dom'
import 'vitest-canvas-mock'

import * as matchers from 'jest-extended'
import { expect } from 'vitest'

import { server } from '../src/mocks/server'

expect.extend(matchers)

export class IntersectionObserver {
  root = null
  rootMargin = ''
  thresholds = []

  disconnect() {
    return null
  }

  observe() {
    return null
  }

  takeRecords() {
    return []
  }

  unobserve() {
    return null
  }
}

class ResizeObserver {
  observe() {
    return null
  }
  unobserve() {
    return null
  }
  disconnect() {
    return null
  }
}

beforeAll(() => {
  global.IntersectionObserver = IntersectionObserver
  window.IntersectionObserver = IntersectionObserver

  global.ResizeObserver = ResizeObserver
  window.ResizeObserver = ResizeObserver

  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation((query) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(), // deprecated
      removeListener: vi.fn(), // deprecated
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  })

  Object.defineProperty(window, 'scrollTo', {
    writable: true,
    value: vi.fn(),
  })

  server.listen({ onUnhandledRequest: 'error' })
})

afterEach(() => server.resetHandlers())

afterAll(() => server.close())
