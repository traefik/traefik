import { Box, Button, Flex, TextField } from '@traefiklabs/faency'
// eslint-disable-next-line import/no-unresolved
import { InputHandle } from '@traefiklabs/faency/dist/components/Input'
import { isUndefined, omitBy } from 'lodash'
import { useCallback, useRef, useState } from 'react'
import { FiSearch, FiXCircle } from 'react-icons/fi'
import { URLSearchParamsInit, useSearchParams } from 'react-router-dom'
import { useDebounceCallback } from 'usehooks-ts'

import IconButton from 'components/buttons/IconButton'

type State = {
  search?: string
  status?: string
  sortBy?: string
  direction?: string
}

export const searchParamsToState = (searchParams: URLSearchParams): State => {
  if (searchParams.size <= 0) return {}

  return omitBy(
    {
      direction: searchParams.get('direction') || undefined,
      search: searchParams.get('search') || undefined,
      sortBy: searchParams.get('sortBy') || undefined,
      status: searchParams.get('status') || undefined,
    },
    isUndefined,
  )
}

type Status = {
  id: string
  value?: string
  name: string
}

const statuses: Status[] = [
  { id: 'all', value: undefined, name: 'All status' },
  { id: 'enabled', value: 'enabled', name: 'Success' },
  { id: 'warning', value: 'warning', name: 'Warnings' },
  { id: 'disabled', value: 'disabled', name: 'Errors' },
]

export const TableFilter = ({ hideStatusFilter }: { hideStatusFilter?: boolean }) => {
  const [searchParams, setSearchParams] = useSearchParams()

  const [state, setState] = useState(searchParamsToState(searchParams))
  const searchInputRef = useRef<InputHandle>(null)

  const onSearch = useDebounceCallback((search?: string) => {
    const newState = omitBy({ ...state, search: search || undefined }, isUndefined)
    setState(newState)
    setSearchParams(newState as URLSearchParamsInit)
  }, 500)

  const onStatusClick = useCallback(
    (status?: string) => {
      const newState = omitBy({ ...state, status: status || undefined }, isUndefined)
      setState(newState)
      setSearchParams(newState as URLSearchParamsInit)
    },
    [setSearchParams, state],
  )

  return (
    <Flex css={{ alignItems: 'center', justifyContent: 'space-between', mb: '$5' }}>
      <Flex>
        {!hideStatusFilter &&
          statuses.map(({ id, value, name }) => (
            <Button
              key={id}
              css={{ marginRight: '$3', boxShadow: 'none' }}
              ghost={state.status !== value}
              variant={state.status !== value ? 'secondary' : 'primary'}
              onClick={() => onStatusClick(value)}
            >
              {name}
            </Button>
          ))}
      </Flex>
      <Box css={{ maxWidth: 200, position: 'relative' }}>
        <TextField
          ref={searchInputRef}
          data-testid="table-search-input"
          defaultValue={state.search || ''}
          onChange={(e) => {
            onSearch(e.target?.value)
          }}
          placeholder="Search"
          css={{ input: { paddingRight: '$6' } }}
          endAdornment={
            state.search ? (
              <IconButton
                type="button"
                css={{ height: '20px', p: 0, color: 'currentColor', '&:before, &:after': { borderRadius: '10px' } }}
                ghost
                icon={<FiXCircle size={20} />}
                onClick={() => {
                  onSearch('')
                  searchInputRef.current?.clear()
                }}
                title="Clear search"
              />
            ) : (
              <FiSearch color="hsl(0, 0%, 56%)" size={20} />
            )
          }
        />
      </Box>
    </Flex>
  )
}

export default TableFilter
