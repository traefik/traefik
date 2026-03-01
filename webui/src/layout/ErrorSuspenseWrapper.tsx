import { ComponentType, ReactNode, Suspense } from 'react'
import { ErrorBoundary, FallbackProps } from 'react-error-boundary'

import ErrorFallback from './ErrorFallback'

type SuspenseWrapperProps = {
  suspenseFallback?: ReactNode
  errorFallback?: ComponentType<FallbackProps>
  silentFail?: boolean
  children?: ReactNode
}

const ErrorSuspenseWrapper = ({
  errorFallback = ErrorFallback,
  suspenseFallback = null,
  silentFail = false,
  children,
}: SuspenseWrapperProps) => {
  return (
    <ErrorBoundary FallbackComponent={silentFail ? () => null : errorFallback}>
      <Suspense fallback={suspenseFallback}>{children}</Suspense>
    </ErrorBoundary>
  )
}

export default ErrorSuspenseWrapper
