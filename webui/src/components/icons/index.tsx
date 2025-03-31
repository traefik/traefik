import { VariantProps } from '@stitches/react'
import { CSS, Flex } from '@traefiklabs/faency'
import { HTMLAttributes } from 'react'

export type CustomIconProps = HTMLAttributes<SVGElement> & {
  color?: string
  fill?: string
  stroke?: string
  width?: number | string
  height?: number | string
  flexProps?: VariantProps<typeof Flex>
  css?: CSS
  viewBox?: string
}
