import { Box, Text } from '@traefiklabs/faency'
import { FallbackProps } from 'react-error-boundary'

import { Button } from 'components/FaencyOverrides'

const ErrorFallback: React.VFC<FallbackProps> = ({ error, resetErrorBoundary }) => {
  return (
    <Box role="alert">
      <Box css={{ mb: '$2' }}>
        <Text as="p">Something went wrong:</Text>
      </Box>
      <Box css={{ mb: '$2' }}>
        <Text variant="red">{error.message}</Text>
      </Box>
      <Button type="button" onClick={resetErrorBoundary}>
        Try again
      </Button>
    </Box>
  )
}

export default ErrorFallback
