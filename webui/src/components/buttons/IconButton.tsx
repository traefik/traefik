import { Button, Flex, Text } from '@traefiklabs/faency'
import { ComponentProps, ReactNode } from 'react'

type IconButtonProps = ComponentProps<typeof Button> & {
  gap?: 1 | 2
  icon: ReactNode
  text?: string
}

export default function IconButton({ css = {}, gap = 2, icon, text, ...props }: IconButtonProps) {
  return (
    <Button variant="primary" size="large" css={{ borderRadius: 0, ...css }} {...props}>
      <Flex align="center" justify="between" gap={gap}>
        {icon}
        {text && <Text css={{ color: 'currentColor', paddingTop: '1px' }}>{text}</Text>}
      </Flex>
    </Button>
  )
}
