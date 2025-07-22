import { Box, CSS } from '@traefiklabs/faency'
import { ReactNode } from 'react'
import { FiAlertCircle, FiAlertTriangle, FiCheckCircle } from 'react-icons/fi'

export type StatusType = 'info' | 'success' | 'warning' | 'error' | 'enabled' | 'disabled'

export const iconByStatus: { [key in StatusType]: ReactNode } = {
  info: <FiAlertCircle color="white" size={20} />,
  success: <FiCheckCircle color="white" size={20} />,
  warning: <FiAlertCircle color="white" size={20} />,
  error: <FiAlertTriangle color="white" size={20} />,
  enabled: <FiCheckCircle color="white" size={20} />,
  disabled: <FiAlertTriangle color="white" size={20} />,
}

// Please notice: dark and light colors have the same values.
export const colorByStatus: { [key in StatusType]: string } = {
  info: 'hsl(220, 67%, 51%)',
  success: '#30A46C',
  warning: 'hsl(24 94.0% 50.0%)',
  error: 'hsl(347, 100%, 60.0%)',
  enabled: '#30A46C',
  disabled: 'hsl(347, 100%, 60.0%)',
}

type StatusProps = {
  css?: CSS
  size?: number
  status: StatusType
}

export default function Status({ css = {}, size = 20, status }: StatusProps) {
  const Icon = ({ size }: { size: number }) => {
    switch (status) {
      case 'info':
        return <FiAlertCircle color="white" size={size} />
      case 'success':
        return <FiCheckCircle color="white" size={size} />
      case 'warning':
        return <FiAlertCircle color="white" size={size} />
      case 'error':
        return <FiAlertTriangle color="white" size={size} />
      case 'enabled':
        return <FiCheckCircle color="white" size={size} />
      case 'disabled':
        return <FiAlertTriangle color="white" size={size} />
      default:
        return null
    }
  }

  return (
    <Box
      css={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        borderRadius: '4px',
        backgroundColor: colorByStatus[status],
        marginRight: '10px',
        padding: '6px',
        ...css,
      }}
    >
      <Icon size={size} />
    </Box>
  )
}
