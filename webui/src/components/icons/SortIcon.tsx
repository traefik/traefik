import { config, Flex } from '@traefiklabs/faency'
import { useEffect, useState } from 'react'

import { CustomIconProps } from 'components/icons'
import { useIsDarkMode } from 'hooks/use-theme'

type SortIconProps = CustomIconProps & {
  order?: 'asc' | 'desc' | ''
}

export default function SortIcon({ css = {}, order, flexProps = {}, ...props }: SortIconProps) {
  const [enabledColor, setEnabledColor] = useState<string>((config.theme.colors as Record<string, string>).deepBlue3)
  const [disabledColor, setDisabledColor] = useState<string>((config.theme.colors as Record<string, string>).deepBlue8)

  const isDarkMode = useIsDarkMode()

  useEffect(() => {
    setEnabledColor((config.theme.colors as Record<string, string>)[isDarkMode ? 'deepBlue3' : 'deepBlue11'])
    setDisabledColor((config.theme.colors as Record<string, string>)[isDarkMode ? 'deepBlue8' : 'deepBlue6'])
  }, [isDarkMode])

  return (
    <Flex {...flexProps} css={css}>
      <svg
        xmlns="http://www.w3.org/2000/svg"
        xmlnsXlink="http://www.w3.org/1999/xlink"
        role="img"
        aria-labelledby="sort-icon"
        viewBox="0 0 8 15"
        {...props}
      >
        <title id="sort-icon">Sort</title>
        <g fill="none" fillRule="evenodd" stroke="none" strokeWidth="1">
          <g transform="translate(-438 -204)">
            <g transform="translate(368 190)">
              <g>
                <path d="M0 0H217V40H0z"></path>
              </g>
              <text
                fill="#000"
                fillOpacity="0.85"
                fontFamily="RubikRoman-Medium, Rubik"
                fontSize="16"
                fontWeight="400"
                letterSpacing="0.615"
              >
                <tspan x="17" y="26.057">
                  Name
                </tspan>
              </text>
            </g>
            <g>
              <g transform="translate(435 201.557)">
                <g transform="translate(3 12)">
                  <path
                    fill={!!order && order === 'desc' ? enabledColor : disabledColor}
                    d="M7.815.2a.72.72 0 010 .964L4.447 4.8a.6.6 0 01-.894 0L.185 1.164a.72.72 0 010-.964.6.6 0 01.893 0L4 3.354 6.922.2a.6.6 0 01.893 0z"
                  ></path>
                </g>
              </g>
              <g transform="translate(435 201.557)">
                <g transform="rotate(180 5.5 4)">
                  <path
                    fill={!!order && order === 'asc' ? enabledColor : disabledColor}
                    d="M7.815.2a.72.72 0 010 .964L4.447 4.8a.6.6 0 01-.894 0L.185 1.164a.72.72 0 010-.964.6.6 0 01.893 0L4 3.354 6.922.2a.6.6 0 01.893 0z"
                  ></path>
                </g>
              </g>
            </g>
          </g>
        </g>
      </svg>
    </Flex>
  )
}
