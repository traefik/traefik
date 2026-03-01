import { AriaTh, CSS, Flex, Label } from '@traefiklabs/faency'
import { useCallback, useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'

import SortButton from 'components/buttons/SortButton'

const STYLE_BY_ALIGN_VALUE = {
  left: {},
  center: {
    justifyContent: 'center',
  },
  right: {
    justifyContent: 'flex-end',
  },
}

type SortableThProps = {
  label: string
  isSortable?: boolean
  sortByValue?: string
  align?: 'left' | 'center' | 'right'
  css?: CSS
}

export default function SortableTh({ label, isSortable = false, sortByValue, align = 'left', css }: SortableThProps) {
  const wrapperStyle = useMemo(() => STYLE_BY_ALIGN_VALUE[align], [align])

  const [searchParams, setSearchParams] = useSearchParams()

  const isActive = useMemo(() => (searchParams.get('sortBy') || 'name') === sortByValue, [searchParams, sortByValue])

  const order = useMemo(() => (searchParams.get('direction') as 'asc' | 'desc' | '') || 'asc', [searchParams])

  const onSort = useCallback(() => {
    if (!sortByValue) return
    const direction = searchParams.get('direction') || 'asc'
    const sortBy = searchParams.get('sortBy') || 'name'
    if (!sortBy || sortBy !== sortByValue || direction === 'desc') {
      setSearchParams({ ...Object.fromEntries(searchParams.entries()), sortBy: sortByValue, direction: 'asc' })
    } else {
      setSearchParams({ ...Object.fromEntries(searchParams.entries()), sortBy: sortByValue, direction: 'desc' })
    }
  }, [sortByValue, searchParams, setSearchParams])

  return (
    <AriaTh css={css}>
      <Flex align="center" css={wrapperStyle}>
        {isSortable ? (
          <SortButton onClick={onSort} order={isActive ? order : undefined} label={label} />
        ) : (
          <Label>{label}</Label>
        )}
      </Flex>
    </AriaTh>
  )
}
