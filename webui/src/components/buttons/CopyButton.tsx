import { Flex, Button, CSS, AccessibleIcon } from '@traefiklabs/faency'
import React, { useState } from 'react'
import { FiCheck, FiCopy } from 'react-icons/fi'

type CopyButtonProps = {
  text: string
  disabled?: boolean
  css?: CSS
  onClick?: () => void
  iconOnly?: boolean
  title?: string
  color?: string
}

const CopyButton = ({
  text,
  disabled,
  css,
  onClick,
  iconOnly = false,
  title = 'Copy',
  color = 'var(--colors-textSubtle)',
}: CopyButtonProps) => {
  const [showConfirmation, setShowConfirmation] = useState(false)

  return (
    <Button
      ghost
      size="small"
      css={{
        color: '$hiContrast',
        px: iconOnly ? '$1' : undefined,
        ...css,
      }}
      title={title}
      onClick={async (e: React.MouseEvent): Promise<void> => {
        e.preventDefault()
        e.stopPropagation()
        await navigator.clipboard.writeText(text)
        if (onClick) onClick()
        setShowConfirmation(true)
        setTimeout(() => setShowConfirmation(false), 1500)
      }}
      disabled={disabled}
      type="button"
    >
      <Flex align="center" gap={2} css={{ userSelect: 'none' }}>
        <AccessibleIcon label="copy">
          {showConfirmation ? <FiCheck color={color} size={14} /> : <FiCopy color={color} size={14} />}
        </AccessibleIcon>
        {!iconOnly ? (showConfirmation ? 'Copied!' : title) : null}
      </Flex>
    </Button>
  )
}

export default CopyButton
