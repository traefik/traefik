import { AriaTr, VariantProps } from '@traefiklabs/faency'
import { ComponentProps, forwardRef, ReactNode } from 'react'
import { useHref } from 'react-router-dom'

type ClickableRowProps = ComponentProps<typeof AriaTr> &
  VariantProps<typeof AriaTr> & {
    children: ReactNode
    to: string
  }

const ClickableRow = forwardRef<HTMLTableRowElement | null, ClickableRowProps>(({ children, to, ...props }, ref) => {
  const href = useHref(to)

  return (
    <AriaTr as="a" href={href} interactive ref={ref} css={{ textDecoration: 'none', ...props.css }} {...props}>
      {children}
    </AriaTr>
  )
})

export default ClickableRow
