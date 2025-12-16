import {
  AriaTable,
  AriaTbody,
  AriaTr,
  CSS,
  AriaTd,
  Flex,
  Skeleton as FaencySkeleton,
  VariantProps,
  AriaThead,
  AriaTh,
} from '@traefiklabs/faency'
import { ReactNode } from 'react'

type AriaTableSkeletonProps = {
  children?: ReactNode
  columns?: number
  css?: CSS
  lastColumnIsNarrow?: boolean
  rowHeight?: string
  rows?: number
  skeletonWidth?: string
}

interface AriaTdSkeletonProps extends VariantProps<typeof FaencySkeleton> {
  css?: CSS
  flexCss?: CSS
}
const AriaTdSkeleton = ({ css = {}, flexCss = {} }: AriaTdSkeletonProps) => (
  <AriaTd css={{ height: 38 }}>
    <Flex css={{ flexDirection: 'column', justifyContent: 'space-around', alignItems: 'flex-start', ...flexCss }}>
      <FaencySkeleton variant="text" css={css} />
    </Flex>
  </AriaTd>
)

const AriaThSkeleton = ({ css = {}, flexCss = {} }: AriaTdSkeletonProps) => (
  <AriaTh css={{ height: 38 }}>
    <Flex css={{ flexDirection: 'column', justifyContent: 'space-around', alignItems: 'flex-start', ...flexCss }}>
      <FaencySkeleton variant="text" css={css} />
    </Flex>
  </AriaTh>
)

export default function AriaTableSkeleton({
  columns = 3,
  css,
  lastColumnIsNarrow = false,
  rowHeight = undefined,
  rows = 5,
  skeletonWidth = '50%',
}: AriaTableSkeletonProps) {
  return (
    <AriaTable css={{ tableLayout: 'auto', ...css }}>
      <AriaThead>
        <AriaTr key="header-row">
          {[...Array(columns)].map((_, colIdx) => (
            <AriaThSkeleton
              key={`header-col-${colIdx}`}
              css={{ width: colIdx === columns - 1 && lastColumnIsNarrow ? '24px' : skeletonWidth }}
            />
          ))}
        </AriaTr>
      </AriaThead>
      <AriaTbody>
        {[...Array(rows)].map((_, rowIdx) => (
          <AriaTr key={`row-${rowIdx}`} css={{ height: rowHeight }}>
            {[...Array(columns)].map((_, colIdx) => (
              <AriaTdSkeleton
                key={`row-${rowIdx}-col-${colIdx}`}
                css={{ width: colIdx === columns - 1 && lastColumnIsNarrow ? '24px' : skeletonWidth }}
              />
            ))}
          </AriaTr>
        ))}
      </AriaTbody>
    </AriaTable>
  )
}
