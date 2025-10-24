import { Flex } from '@traefiklabs/faency'
import { useId } from 'react'

import { CustomIconProps } from 'components/icons'

const GatewayIcon = ({ color = 'currentColor', css = {}, ...props }: CustomIconProps) => {
  const titleId = useId()

  return (
    <Flex css={css}>
      <svg
        width="48px"
        height="48px"
        viewBox="0 0 48 48"
        version="1.1"
        xmlns="http://www.w3.org/2000/svg"
        xmlnsXlink="http://www.w3.org/1999/xlink"
        {...props}
      >
        <title id={titleId}>gateways_icon</title>
        <defs>
          <linearGradient x1="50%" y1="100%" x2="50%" y2="-30.9671875%" id="linearGradient-1">
            <stop stopColor={color} stopOpacity="0.322497815" offset="0%"></stop>
            <stop stopColor={color} stopOpacity="0" offset="100%"></stop>
          </linearGradient>
          <linearGradient x1="50%" y1="100%" x2="50%" y2="-30.9671875%" id="linearGradient-2">
            <stop stopColor={color} stopOpacity="0.322497815" offset="0%"></stop>
            <stop stopColor={color} stopOpacity="0" offset="100%"></stop>
          </linearGradient>
          <linearGradient x1="100%" y1="50%" x2="-5.3329563e-11%" y2="50%" id={`linearGradient-3-${titleId}`}>
            <stop stopColor={color} offset="0%"></stop>
            <stop stopColor={color} stopOpacity="0.2" offset="100%"></stop>
          </linearGradient>
        </defs>
        <g id={titleId} stroke="none" strokeWidth="1" fill="none" fillRule="evenodd">
          <g id="Dashboard-Component-States" transform="translate(-819, -59)">
            <g id="Icon" transform="translate(823, 67)">
              <path
                d="M11.8810613,8.00728271 L11.8810613,21.2735224 L20.1151876,25.72 L28.3494968,21.2735224 L28.3494968,8.00728271 C28.3494968,8.00728271 25.6047575,5.72 20.115279,5.72 C14.6258005,5.72 11.8810613,8.00728271 11.8810613,8.00728271 Z"
                id="Rectangle-37"
                fill="url(#linearGradient-2)"
              ></path>
              <path
                d="M15.1747484,11.0923696 L15.1747484,19.0521134 L20.1152242,21.72 L25.0558097,19.0521134 L25.0558097,11.0923696 C25.0558097,11.0923696 23.4089661,9.72 20.115279,9.72 C16.8215919,9.72 15.1747484,11.0923696 15.1747484,11.0923696 Z"
                id="Rectangle-37"
                fill={color}
              ></path>
              <path
                d="M10.65,2.32629257 L10.65,29.6737074 L1.35,24.3594217 L1.35,7.64057828 L10.65,2.32629257 Z"
                id="Rectangle"
                stroke={`url(#linearGradient-3-${titleId})`}
                strokeWidth="2.7"
              ></path>
              <path
                d="M38.65,2.32629257 L38.65,29.6737074 L29.35,24.3594217 L29.35,7.64057828 L38.65,2.32629257 Z"
                id="Rectangle"
                stroke={`url(#linearGradient-3-${titleId})`}
                strokeWidth="2.7"
                transform="translate(34, 16) scale(-1, 1) translate(-34, -16)"
              ></path>
            </g>
          </g>
        </g>
      </svg>
    </Flex>
  )
}

export default GatewayIcon
