export type PositionXProps = 'left' | 'center' | 'right'
export type PositionYProps = 'top' | 'bottom'

export type GetPositionType = {
  top?: number
  bottom?: number
  left?: number
  right?: number
}

export function getPositionValues(positionX: PositionXProps, positionY: PositionYProps): GetPositionType {
  const position: GetPositionType = {}

  switch (positionX) {
    case 'left':
      position.left = 0
      break
    case 'center':
      position.left = 0
      position.right = 0
      break
    case 'right':
      position.right = 0
      break
  }

  switch (positionY) {
    case 'top':
      position.top = 0
      break
    case 'bottom':
      position.bottom = 0
      break
  }

  return position
}
