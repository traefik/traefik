/*
Copyright (C) 2022-2024 Traefik Labs
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.
You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

import { Button, Flex, Text } from '@traefiklabs/faency'
import { ComponentProps, ReactNode } from 'react'

type IconButtonProps = ComponentProps<typeof Button> & {
  gap?: 1 | 2
  icon: ReactNode
  text?: string
}

export default function IconButton({ css = {}, gap = 2, icon, text, ...props }: IconButtonProps) {
  return (
    <Button variant="primary" size="large" css={{ borderRadius: 0, ...css }} {...props}>
      <Flex align="center" justify="between" gap={gap}>
        {icon}
        {text && <Text css={{ color: 'currentColor', paddingTop: '1px' }}>{text}</Text>}
      </Flex>
    </Button>
  )
}
