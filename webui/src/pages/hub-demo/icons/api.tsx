import { Flex } from '@traefiklabs/faency'
import { useId } from 'react'

import { CustomIconProps } from 'components/icons'

const ApiIcon = ({ color = 'currentColor', css = {}, ...props }: CustomIconProps) => {
  const linearGradient1Id = useId()
  const linearGradient2Id = useId()
  const linearGradient3Id = useId()
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
        <title id={titleId}>apis</title>
        <defs>
          <linearGradient x1="71.828%" y1="92.873%" x2="35.088%" y2="3.988%" id={linearGradient1Id}>
            <stop stopColor={color} offset="0%" />
            <stop stopColor={color} stopOpacity="0" offset="100%" />
          </linearGradient>
          <linearGradient x1="64.659%" y1="7.289%" x2="29.649%" y2="91.224%" id={linearGradient2Id}>
            <stop stopColor={color} offset="0%" />
            <stop stopColor={color} stopOpacity="0" offset="100%" />
          </linearGradient>
          <linearGradient x1="10.523%" y1="76.967%" x2="87.277%" y2="76.967%" id={linearGradient3Id}>
            <stop stopColor={color} offset="0%" />
            <stop stopColor={color} stopOpacity="0" offset="100%" />
          </linearGradient>
        </defs>
        <g fill="none" fillRule="evenodd">
          <path
            d="M12.425 19.85a7.425 7.425 0 1 0 0-14.85 7.425 7.425 0 0 0 0 14.85z"
            stroke={color}
            strokeWidth="1.35"
            opacity=".7"
            strokeLinejoin="round"
          />
          <path
            d="M3.375 7.425c0-.411.061-.809.175-1.183l-1.22-.61a5.401 5.401 0 0 0 4.42 7.151V11.42a4.051 4.051 0 0 1-3.375-3.994z"
            fill={`url(#${linearGradient1Id})`}
            transform="translate(5 5)"
          />
          <path
            d="M8.1 11.419v1.364a5.401 5.401 0 0 0 4.42-7.15l-1.22.61a4.051 4.051 0 0 1-3.2 5.177z"
            fill={`url(#${linearGradient2Id})`}
            transform="translate(5 5)"
          />
          <path
            d="M7.425 3.375a4.044 4.044 0 0 0-3.27 1.66l-1.22-.61a5.395 5.395 0 0 1 4.49-2.4c1.872 0 3.522.953 4.49 2.4l-1.22.61a4.044 4.044 0 0 0-3.27-1.66z"
            fill={`url(#${linearGradient3Id})`}
            transform="translate(5 5)"
          />
          <path d="M12.425 14.45a2.025 2.025 0 1 0 0-4.05 2.025 2.025 0 0 0 0 4.05z" fill={color} fillRule="nonzero" />
        </g>
      </svg>
    </Flex>
  )
}

export default ApiIcon
