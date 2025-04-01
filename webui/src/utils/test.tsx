import { cleanup, render } from '@testing-library/react'
import { FaencyProvider } from '@traefiklabs/faency'
import { HelmetProvider } from 'react-helmet-async'
import { BrowserRouter } from 'react-router-dom'
import { SWRConfig } from 'swr'
import { afterEach } from 'vitest'

import fetch from '../libs/fetch'

afterEach(() => {
  cleanup()
})

function customRender(ui: React.ReactElement, options = {}) {
  return render(ui, {
    // wrap provider(s) here if needed
    wrapper: ({ children }) => children,
    ...options,
  })
}

// eslint-disable-next-line import/export
export * from '@testing-library/react'
export { default as userEvent } from '@testing-library/user-event'
// override render export
export { customRender as render } // eslint-disable-line import/export

export function renderWithProviders(ui: React.ReactElement) {
  return customRender(ui, {
    wrapper: ({ children }) => (
      <FaencyProvider>
        <HelmetProvider>
          <SWRConfig
            value={{
              revalidateOnFocus: false,
              fetcher: fetch,
            }}
          >
            <BrowserRouter>{children}</BrowserRouter>
          </SWRConfig>
        </HelmetProvider>
      </FaencyProvider>
    ),
  })
}
