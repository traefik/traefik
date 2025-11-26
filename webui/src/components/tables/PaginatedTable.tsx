import { AriaTable, AriaTbody, AriaTd, AriaTh, AriaThead, AriaTr, Box, Button, Flex, Text } from '@traefiklabs/faency'
import { ReactNode, useEffect, useRef, useState } from 'react'
import { FiChevronsLeft, FiChevronsRight } from 'react-icons/fi'

type PaginatedTableProps<T extends Record<string, unknown>> = {
  data: T[]
  columns: {
    key: keyof T
    header: string
  }[]
  itemsPerPage?: number
  testId?: string
  renderCell?: (key: keyof T, value: T[keyof T], row: T) => ReactNode
}

const PaginatedTable = <T extends Record<string, unknown>>({
  data,
  columns,
  itemsPerPage = 5,
  testId,
  renderCell,
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

  const handlePreviousPage = () => {
    setCurrentPage((prev) => Math.max(0, prev - 1))
  }

  const handleNextPage = () => {
    setCurrentPage((prev) => Math.min(totalPages - 1, prev + 1))
  }

  const getCellContent = (key: keyof T, value: T[keyof T], row: T) => {
    if (renderCell) {
      return renderCell(key, value, row)
    }
    return value as ReactNode
  }

  return (
    <Box>
      <AriaTable
        ref={tableRef}
        data-testid={testId}
        css={totalPages > 1 && tableHeight ? { minHeight: `${tableHeight}px` } : undefined}
      >
        <AriaThead>
          <AriaTr>
            {columns.map((column) => (
              <AriaTh key={String(column.key)}>{column.header}</AriaTh>
            ))}
          </AriaTr>
        </AriaThead>
        <AriaTbody css={totalPages > 1 && tableHeight ? { verticalAlign: 'top' } : undefined}>
          {currentData.map((row, rowIndex) => (
            <AriaTr key={rowIndex}>
              {columns.map((column) => (
                <AriaTd key={String(column.key)}>{getCellContent(column.key, row[column.key], row)}</AriaTd>
              ))}
            </AriaTr>
          ))}
        </AriaTbody>
      </AriaTable>
      {totalPages > 1 && (
        <Flex justify="center" align="center" gap={2} css={{ mt: '$1' }}>
          <Button ghost onClick={handlePreviousPage} disabled={currentPage === 0}>
            <FiChevronsLeft />
          </Button>
          <Text css={{ fontSize: '14px', color: '$textSubtle' }}>
            Page {currentPage + 1} of {totalPages}
          </Text>
          <Button ghost onClick={handleNextPage} disabled={currentPage === totalPages - 1}>
            <FiChevronsRight />
          </Button>
        </Flex>
      )}
    </Box>
  )
}

export default PaginatedTable
