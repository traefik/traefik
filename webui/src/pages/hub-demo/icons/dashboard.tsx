import { Flex } from '@traefiklabs/faency'

import { CustomIconProps } from 'components/icons'

const DashboardIcon = ({ color = 'currentColor', css = {}, ...props }: CustomIconProps) => {
  return (
    <Flex css={css}>
      <svg
        width="24"
        height="24"
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
        role="img"
        aria-labelledby="dashboard-icon"
        {...props}
      >
        <title id="dashboard-icon">dashboard</title>
        <g stroke={color} strokeWidth="1.35" fill="none" fillRule="evenodd" strokeLinejoin="round">
          <path opacity=".7" d="M11.075 5H5v6.075h6.075z" />
          <path d="M19.85 5h-6.075v6.075h6.075zM11.075 13.775H5v6.075h6.075z" />
          <path opacity=".7" d="M19.85 13.775h-6.075v6.075h6.075z" />
        </g>
      </svg>
    </Flex>
  )
}

export default DashboardIcon
