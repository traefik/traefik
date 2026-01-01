import { Box, Button, Flex, H1, Text } from '@traefiklabs/faency'
import { useNavigate } from 'react-router-dom'

import PageTitle from 'layout/PageTitle'

export const NotFound = () => {
  const navigate = useNavigate()

  return (
    <Flex css={{ flexDirection: 'column', alignItems: 'center', p: '$6' }} data-testid="Not found page">
      <PageTitle title="Not found" />
      <Box>
        <H1 style={{ fontSize: '80px', lineHeight: '120px' }}>404</H1>
      </Box>
      <Box css={{ pb: '$3' }}>
        <Text size={6}>I&apos;m sorry, nothing around here...</Text>
      </Box>
      <Button variant="primary" onClick={() => navigate(-1)}>
        Go back
      </Button>
    </Flex>
  )
}
