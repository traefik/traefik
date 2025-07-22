import { waitFor } from '@testing-library/react'
import { useContext, useEffect } from 'react'

import { ToastPool } from './ToastPool'

import { ToastContext, ToastProvider } from 'contexts/toasts'
import { renderWithProviders } from 'utils/test'

describe('<ToastPool />', () => {
  it('should render the toast pool', () => {
    renderWithProviders(<ToastPool />)
  })

  it('should render toasts from context', async () => {
    const Component = () => {
      const { addToast } = useContext(ToastContext)

      useEffect(() => {
        addToast({
          message: 'Test 1',
          severity: 'success',
        })
      }, [addToast])

      return <ToastPool />
    }

    const { getByTestId } = renderWithProviders(
      <ToastProvider>
        <Component />
      </ToastProvider>,
    )

    await waitFor(() => getByTestId('toast-pool'))

    const toastPool = getByTestId('toast-pool')
    expect(toastPool.childNodes.length).toBe(1)
    expect(toastPool.innerHTML).toContain('Test 1')
  })

  it('should render all valid severities of toasts', async () => {
    const Component = () => {
      const { addToast } = useContext(ToastContext)

      useEffect(() => {
        addToast({
          message: 'Test 2',
          severity: 'error',
        })

        addToast({
          message: 'Test 3',
          severity: 'warning',
        })

        addToast({
          message: 'Test 4',
          severity: 'info',
        })
      }, [addToast])

      return <ToastPool />
    }

    const { getByTestId } = renderWithProviders(
      <ToastProvider>
        <Component />
      </ToastProvider>,
    )

    await waitFor(() => getByTestId('toast-pool'))

    const toastPool = getByTestId('toast-pool')
    expect(toastPool.childNodes.length).toBe(3)
    expect(toastPool.innerHTML).toContain('Test 2')
    expect(toastPool.innerHTML).toContain('Test 3')
    expect(toastPool.innerHTML).toContain('Test 4')
  })
})
