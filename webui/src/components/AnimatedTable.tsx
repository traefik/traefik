import { Table, Tbody, Tr } from '@traefiklabs/faency'
import { motion } from 'framer-motion'
import { ReactNode } from 'react'

import ClickableRow from './ClickableRow'

const tableVariants = {
  hidden: {
    opacity: 0,
  },
  visible: {
    opacity: 1,
  },
}

const bodyVariants = {
  hidden: {
    x: -6,
    opacity: 0,
  },
  visible: {
    x: 0,
    opacity: 1,
    transition: { staggerChildren: 0.03, delayChildren: 0.1 },
  },
}

const rowVariants = {
  hidden: {
    x: -6,
    opacity: 0,
  },
  visible: {
    x: 0,
    opacity: 1,
  },
}

const CustomTable = motion.create(Table)
const CustomBody = motion.create(Tbody)
const CustomClickableRow = motion.create(ClickableRow)

const CustomRow = motion.create(Tr)

type AnimatedTableProps = {
  children: ReactNode
  isMounted?: boolean
}

export const AnimatedTable = ({ isMounted = true, children }: AnimatedTableProps) => (
  <CustomTable
    style={{ tableLayout: 'auto' }}
    initial="hidden"
    animate={isMounted ? 'visible' : 'hidden'}
    variants={tableVariants}
  >
    {children}
  </CustomTable>
)

type AnimatedTBodyProps = {
  pageCount: number
  isMounted: boolean
  children: ReactNode | null
}

export const AnimatedTBody = ({ pageCount, isMounted = true, children }: AnimatedTBodyProps) => (
  <CustomBody variants={bodyVariants} animate={isMounted && pageCount > 0 ? 'visible' : 'hidden'}>
    {children}
  </CustomBody>
)

type AnimatedRowType = {
  children: ReactNode
  to?: string
}

export const AnimatedRow = ({ children, to }: AnimatedRowType) => {
  if (to) {
    return (
      <CustomClickableRow to={to} variants={rowVariants}>
        {children}
      </CustomClickableRow>
    )
  }

  return <CustomRow variants={rowVariants}>{children}</CustomRow>
}
