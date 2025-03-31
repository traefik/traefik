import { Flex } from '@traefiklabs/faency'
import { ToastContext } from 'contexts/toasts'
import { useContext } from 'react'
import { getPositionValues, PositionXProps, PositionYProps } from 'utils/position'

import { Toast } from './Toast'

export type ToastPoolProps = {
  positionX?: PositionXProps
  positionY?: PositionYProps
  toastTimeout?: number
}

export const ToastPool = ({ positionX = 'right', positionY = 'bottom', toastTimeout = 5000 }: ToastPoolProps) => {
  const { toasts, hideToast } = useContext(ToastContext)

  return (
    <Flex
      {...getPositionValues(positionX, positionY)}
      css={{
        position: 'fixed',
        bottom: 0,
        flexDirection: 'column',
        maxWidth: '380px',
        zIndex: 2,
        px: '$3',
        margin: positionX === 'center' ? 'auto' : 0,
      }}
      data-testid="toast-pool"
    >
      {toasts?.map((toast) => (
        <Toast key={toast.key} {...toast} dismiss={(): void => hideToast(toast)} timeout={toastTimeout} />
      ))}
    </Flex>
  )
}
