import { AriaTd, Flex, Text } from '@traefiklabs/faency'
import { FiAlertTriangle } from 'react-icons/fi'

type EmptyPlaceholderProps = {
  message?: string
}
export const EmptyPlaceholder = ({ message = 'No data available' }: EmptyPlaceholderProps) => (
  <Flex align="center" justify="center" css={{ py: '$5', color: '$primary' }}>
    <FiAlertTriangle size={16} />
    <Text css={{ pl: '$2' }}>{message}</Text>
  </Flex>
)

export const EmptyPlaceholderTd = (props: EmptyPlaceholderProps) => {
  return (
    <AriaTd css={{ pointerEvents: 'none' }} fullColSpan>
      <EmptyPlaceholder {...props} />
    </AriaTd>
  )
}
