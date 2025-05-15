import { Tr, VariantProps } from '@traefiklabs/faency'
import { ComponentProps, forwardRef, KeyboardEvent, MouseEvent, ReactNode, useCallback } from 'react'
import { useHref, useNavigate } from 'react-router-dom'

type ClickableRowProps = ComponentProps<typeof Tr> &
  VariantProps<typeof Tr> & {
    children: ReactNode
    to: string
  }

const ClickableRow = forwardRef<HTMLTableRowElement | null, ClickableRowProps>(({ children, to, ...props }, ref) => {
  const navigate = useNavigate()
  const href = useHref(to)

  const onClick = useCallback(
    (event: MouseEvent<HTMLTableRowElement>) => {
      if (event.ctrlKey || event.metaKey) {
        window.open(href, '_blank', 'noopener,noreferrer')
        return
      }

      event.preventDefault()
      navigate(to)
    },
    [href, navigate, to],
  )

  const onKeyDown = (event: KeyboardEvent<HTMLTableRowElement>) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      navigate(to)
    }
  }

  return (
    <Tr
      ref={ref}
      role="link"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={onKeyDown}
      interactive
      css={{ cursor: 'pointer', ...props.css }}
      {...props}
    >
      {children}
    </Tr>
  )
})

export default ClickableRow
