import { Box, Button, Flex, styled, Text } from '@traefiklabs/faency'
import { AnimatePresence, motion } from 'framer-motion'
import { ReactNode, useEffect } from 'react'
import { FiX } from 'react-icons/fi'

import { colorByStatus, iconByStatus, StatusType } from 'components/resources/Status'

const CloseButton = styled(Button, {
  position: 'absolute',
  top: 0,
  right: 0,
  padding: 0,
})

const ToastContainer = styled(Flex, {
  marginBottom: '$3',
  width: '100%',
  position: 'relative',
  padding: '$5 $6',
  borderRadius: '$2',
})

const AnimatedToastContainer = motion.create(ToastContainer)

const toastVariants = {
  create: {
    opacity: 0,
    y: 100,
  },
  visible: {
    opacity: 1,
    y: 0,
  },
  hidden: {
    opacity: 0,
    x: '100%',
    scale: 0,
  },
}

export type ToastState = {
  severity: StatusType
  message?: string
  isVisible?: boolean
  key?: string
}

type ToastProps = ToastState & {
  dismiss: () => void
  icon?: ReactNode
  timeout?: number
}

export const Toast = ({ message, dismiss, severity = 'error', icon, isVisible = true, timeout = 0 }: ToastProps) => {
  useEffect(() => {
    if (timeout) {
      setTimeout(() => dismiss(), timeout)
    }
  }, [timeout, dismiss])

  const propsBySeverity = {
    info: {
      color: colorByStatus.info,
      icon: iconByStatus.info,
    },
    success: {
      color: colorByStatus.success,
      icon: iconByStatus.success,
    },
    warning: {
      color: colorByStatus.warning,
      icon: iconByStatus.warning,
    },
    error: {
      color: colorByStatus.error,
      icon: iconByStatus.error,
    },
  }

  return (
    <AnimatePresence>
      {isVisible && (
        <AnimatedToastContainer
          css={{ backgroundColor: propsBySeverity[severity].color }}
          gap={2}
          initial="create"
          animate="visible"
          exit="hidden"
          variants={toastVariants}
        >
          <Box css={{ width: '$4', height: '$4' }}>{icon ? icon : propsBySeverity[severity].icon}</Box>
          <Text css={{ color: 'white', fontWeight: 600, lineHeight: '$4' }}>{message}</Text>
          {!timeout && (
            <CloseButton ghost onClick={dismiss} css={{ px: '$2' }}>
              <FiX color="#fff" size={20} />
            </CloseButton>
          )}
        </AnimatedToastContainer>
      )}
    </AnimatePresence>
  )
}
