import { Box, Button, Flex, H1, Text } from '@traefiklabs/faency'
import { useNavigate } from 'react-router-dom'

import Page from 'layout/Page'

export const NotFound = () => {
  const navigate = useNavigate()

  return (
    <Page title="Not found">
      <Flex css={{ flexDirection: 'column', alignItems: 'center', p: '$6' }}>
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
    </Page>
  )
}
