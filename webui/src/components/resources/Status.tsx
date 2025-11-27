import { Box, CSS } from '@traefiklabs/faency'
import { ReactNode } from 'react'
import { FiAlertCircle, FiAlertTriangle, FiCheckCircle, FiLoader } from 'react-icons/fi'

export const iconByStatus: { [key in Resource.Status]: ReactNode } = {
  info: <FiAlertCircle color="currentColor" size={20} />,
  success: <FiCheckCircle color="currentColor" size={20} />,
  warning: <FiAlertCircle color="currentColor" size={20} />,
  error: <FiAlertTriangle color="currentColor" size={20} />,
  enabled: <FiCheckCircle color="currentColor" size={20} />,
  disabled: <FiAlertTriangle color="currentColor" size={20} />,
  loading: <FiLoader color="currentColor" size={20} />,
}

// Please notice: dark and light colors have the same values.
export const colorByStatus: { [key in Resource.Status]: string } = {
  info: 'hsl(220, 67%, 51%)',
  success: '#30A46C',
  warning: 'hsl(24 94.0% 50.0%)',
  error: 'hsl(347, 100%, 60.0%)',
  enabled: '#30A46C',
  disabled: 'hsl(347, 100%, 60.0%)',
  loading: 'hsla(0, 0%, 100%, 0.51)',
}

type StatusProps = {
  css?: CSS
  size?: number
  status: Resource.Status
  color?: string
}

export default function Status({ css = {}, size = 20, status, color = 'white' }: StatusProps) {
  const Icon = ({ size }: { size: number }) => {
    switch (status) {
      case 'info':
        return <FiAlertCircle color={color} size={size} />
      case 'success':
        return <FiCheckCircle color={color} size={size} />
      case 'warning':
        return <FiAlertCircle color={color} size={size} />
      case 'error':
        return <FiAlertTriangle color={color} size={size} />
      case 'enabled':
        return <FiCheckCircle color={color} size={size} />
      case 'disabled':
        return <FiAlertTriangle color={color} size={size} />
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
