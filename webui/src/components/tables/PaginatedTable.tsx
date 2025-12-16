import { AriaTable, AriaTbody, AriaTd, AriaThead, AriaTr, Box, Button, Flex, Text } from '@traefiklabs/faency'
import { ReactNode, useEffect, useRef, useState } from 'react'
import { FiChevronLeft, FiChevronRight, FiChevronsLeft, FiChevronsRight } from 'react-icons/fi'

import SortableTh from './SortableTh'

type PaginatedTableProps<T extends Record<string, unknown>> = {
  data: T[]
  columns: {
    key: keyof T
    header: string
    sortable?: boolean
    width?: string
  }[]
  itemsPerPage?: number
  testId?: string
  renderCell?: (key: keyof T, value: T[keyof T], row: T) => ReactNode
  renderRow?: (row: T) => ReactNode
}

const PaginatedTable = <T extends Record<string, unknown>>({
  data,
  columns,
  itemsPerPage = 5,
  testId,
  renderCell,
  renderRow,
}: PaginatedTableProps<T>) => {
  const [currentPage, setCurrentPage] = useState(0)
  const [tableHeight, setTableHeight] = useState<number | undefined>(undefined)
  const tableRef = useRef<HTMLTableSectionElement>(null)

  const totalPages = Math.ceil(data.length / itemsPerPage)
  const startIndex = currentPage * itemsPerPage
  const endIndex = startIndex + itemsPerPage
  const currentData = data.slice(startIndex, endIndex)

  // Workaround to keep the same height to avoid layout shift
  useEffect(() => {
    if (totalPages > 1 && currentPage === 0 && tableRef.current && !tableHeight) {
      const height = tableRef.current.offsetHeight
      setTableHeight(height)
    }
  }, [totalPages, currentPage, tableHeight])

  const handleFirstPage = () => {
    setCurrentPage(0)
  }

  const handlePreviousPage = () => {
    setCurrentPage((prev) => Math.max(0, prev - 1))
  }

  const handleNextPage = () => {
    setCurrentPage((prev) => Math.min(totalPages - 1, prev + 1))
  }

  const handleLastPage = () => {
    setCurrentPage(totalPages - 1)
  }

  const getCellContent = (key: keyof T, value: T[keyof T], row: T) => {
    if (renderCell) {
      return renderCell(key, value, row)
    }
    return value as ReactNode
  }

  return (
    <Box>
      <AriaTable ref={tableRef} css={totalPages > 1 && tableHeight ? { minHeight: `${tableHeight}px` } : undefined}>
        <AriaThead>
          <AriaTr>
            {columns.map((column) => (
              <SortableTh
                key={String(column.key)}
                label={column.header}
                isSortable={column.sortable}
                sortByValue={column.sortable ? String(column.key) : undefined}
                css={column.width ? { width: column.width } : undefined}
              />
            ))}
          </AriaTr>
        </AriaThead>

        <AriaTbody data-testid={testId} css={totalPages > 1 && tableHeight ? { verticalAlign: 'top' } : undefined}>
          {currentData.map((row, rowIndex) => {
            if (renderRow) {
              return renderRow(row)
            }

            const rowContent = (
              <>
                {columns?.map((column) => (
                  <AriaTd key={String(column.key)}>{getCellContent(column.key, row[column.key], row)}</AriaTd>
                ))}
              </>
            )

            return <AriaTr key={rowIndex}>{rowContent}</AriaTr>
          })}
        </AriaTbody>
      </AriaTable>
      {totalPages > 1 && (
        <Flex justify="center" align="center" gap={2} css={{ mt: '$1' }}>
          <Flex>
            <Button
              ghost
              onClick={handleFirstPage}
              disabled={currentPage === 0}
              aria-label="Go to first page"
              css={{ px: '$1' }}
            >
              <FiChevronsLeft aria-label="First page" />
            </Button>
            <Button
              ghost
              onClick={handlePreviousPage}
              disabled={currentPage === 0}
              aria-label="Go to previous page"
              css={{ px: '$1' }}
            >
              <FiChevronLeft aria-label="Previous page" />
            </Button>
          </Flex>
          <Text css={{ fontSize: '14px', color: '$textSubtle' }}>
            Page {currentPage + 1} of {totalPages}
          </Text>
          <Button
            ghost
            onClick={handleNextPage}
            disabled={currentPage === totalPages - 1}
            aria-label="Go to next page"
            css={{ px: '$1' }}
          >
            <FiChevronRight aria-label="Next page" />
          </Button>
          <Button
            ghost
            onClick={handleLastPage}
            disabled={currentPage === totalPages - 1}
            aria-label="Go to last page"
            css={{ px: '$1' }}
          >
            <FiChevronsRight aria-label="Last page" />
          </Button>
        </Flex>
      )}
    </Box>
  )
}

export default PaginatedTable
