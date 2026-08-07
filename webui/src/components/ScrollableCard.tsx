import { Card, styled } from '@traefik-labs/faency'

const ScrollableCard = styled(Card, {
  width: '100%',
  maxHeight: 300,
  overflowY: 'auto',
  overflowX: 'hidden',
  scrollbarWidth: 'thin',
  scrollbarColor: '$colors-primary $colors-01dp',
  scrollbarGutter: 'stable',
})

export default ScrollableCard
