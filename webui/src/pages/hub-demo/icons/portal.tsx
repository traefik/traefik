import { Flex } from '@traefiklabs/faency'
import { useId } from 'react'

import { CustomIconProps } from 'components/icons'

const PortalIcon = ({ color = 'currentColor', css = {}, ...props }: CustomIconProps) => {
  const linearGradientId = useId()
  const titleId = useId()

  return (
    <Flex css={css}>
      <svg
        xmlns="http://www.w3.org/2000/svg"
        aria-labelledby={titleId}
        width="24"
        height="24"
        viewBox="0 0 24 24"
        role="img"
        {...props}
      >
        <title id={titleId}>portals</title>
        <defs>
          <linearGradient x1="50%" y1="17.22%" x2="50%" y2="100%" id={linearGradientId}>
            <stop stopColor={color} offset="0%" />
            <stop stopColor={color} stopOpacity=".468" offset="100%" />
          </linearGradient>
        </defs>
        <g fill="none" fillRule="evenodd">
          <path
            stroke={`url(#${linearGradientId})`}
            strokeWidth="1.35"
            strokeLinejoin="round"
            d="M14.85 0H0v13.5h14.85z"
            transform="translate(5 5)"
          />
          <path
            d="M7.7 7.7a.675.675 0 1 0 0-1.35.675.675 0 0 0 0 1.35zM10.4 7.7a.675.675 0 1 0 0-1.35.675.675 0 0 0 0 1.35z"
            fill={color}
            fillRule="nonzero"
          />
          <path stroke={color} strokeWidth="1.35" strokeLinejoin="round" d="M5 9.05h14.85" />
        </g>
      </svg>
    </Flex>
  )
}

export default PortalIcon
