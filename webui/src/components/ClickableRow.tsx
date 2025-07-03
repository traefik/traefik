import { AriaTr, VariantProps, styled } from '@traefiklabs/faency'
import { ComponentProps, forwardRef, ReactNode } from 'react'
import { useHref } from 'react-router-dom'

const UnstyledLink = styled('a', {
  color: 'inherit',
  textDecoration: 'inherit',
  fontWeight: 'inherit',
  '&:hover': {
    cursor: 'pointer',
  },
})

type ClickableRowProps = ComponentProps<typeof AriaTr> &
  VariantProps<typeof AriaTr> & {
    children: ReactNode
    to: string
  }

export default forwardRef<HTMLTableRowElement | null, ClickableRowProps>(({ children, css, to, ...props }, ref) => {
  const href = useHref(to)

  return (
    <AriaTr asChild interactive ref={ref} css={css} {...props}>
      <UnstyledLink href={href}>{children}</UnstyledLink>
    </AriaTr>
  )
})
