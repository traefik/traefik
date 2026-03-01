import { styled, Flex, Label } from '@traefiklabs/faency'
import { ComponentProps } from 'react'

import SortIcon from 'components/icons/SortIcon'

const StyledSortButton = styled('button', {
  border: 'none',
  margin: 0,
  padding: 0,
  overflow: 'visible',
  background: 'transparent',
  color: 'inherit',
  font: 'inherit',
  verticalAlign: 'middle',
  lineHeight: 'normal',
  '-webkit-font-smoothing': 'inherit', // @FIXME not on standard tracks https://developer.mozilla.org/en-US/docs/Web/CSS/font-smooth
  '-moz-osx-font-smoothing': 'inherit',
  '-webkit-appearance': 'none',
  '&:focus': {
    outline: 0,
    color: '$hiContrast',
  },
  '&::-moz-focus-inner': {
    // @FIXME not on standard tracks https://developer.mozilla.org/en-US/docs/Web/CSS/::-moz-focus-inner
    border: 0,
    padding: 0,
  },
  '@hover': {
    '&:hover': {
      cursor: 'pointer',
      color: '$hiContrast',
    },
  },
})

export default function SortButton({
  label,
  order,
  ...props
}: ComponentProps<typeof StyledSortButton> & { order?: 'asc' | 'desc' | ''; label?: string }) {
  return (
    <StyledSortButton type="button" {...props}>
      <Flex align="center">
        {label && <Label css={{ cursor: 'inherit', color: 'inherit' }}>{label}</Label>}
        <SortIcon height={15} css={{ ml: '$2' }} order={order} />
      </Flex>
    </StyledSortButton>
  )
}
