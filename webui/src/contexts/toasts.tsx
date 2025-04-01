import * as React from 'react'
import * as sh from 'shorthash'

import { ToastState } from 'components/Toast'

function handleHideToast(toast: ToastState): (t: ToastState) => ToastState {
  return (t: ToastState): ToastState => {
    if (t === toast) {
      t.isVisible = false
    }
    return t
  }
}

interface ToastProviderProps {
  children: React.ReactNode
}

interface ToastContextProps {
  toasts: ToastState[]
  addToast: (toast: ToastState) => void
  hideToast: (toast: ToastState) => void
}

export const ToastContext = React.createContext({} as ToastContextProps)

export const ToastProvider = (props: ToastProviderProps) => {
  const [toasts, setToastList] = React.useState<ToastState[]>([])
  const addToast = React.useCallback(
    (toast: ToastState) => {
      const key = sh.unique(JSON.stringify(toast))
      if (!toasts.find((t) => t.key === key)) {
        toast.key = key
        setToastList((toasts) => [...toasts, toast])
      }
    },
    [setToastList, toasts],
  )

  const hideToast = React.useCallback(
    (toast: ToastState) => {
      setToastList((toasts) => toasts.map(handleHideToast(toast)))
    },
    [setToastList],
  )

  const value: ToastContextProps = { toasts, addToast, hideToast }

  return <ToastContext.Provider value={value}>{props.children}</ToastContext.Provider>
}
