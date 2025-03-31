import { Flex, Text } from '@traefiklabs/faency'
import { FiAlertTriangle } from 'react-icons/fi'

export const EmptyPlaceholder = ({ message = 'No data available' }: { message?: string }) => (
  <Flex align="center" justify="center" css={{ py: '$5' }}>
    <FiAlertTriangle size={16} color="hsl(222, 67%, 51%)" />
    <Text css={{ pl: '$2' }}>{message}</Text>
  </Flex>
)
