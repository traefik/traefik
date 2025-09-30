import { Flex } from '@traefiklabs/faency'
import { motion } from 'framer-motion'
import { FiLoader } from 'react-icons/fi'

export const SpinnerLoader = ({ size = 24 }: { size?: number }) => (
  <motion.div
    animate={{
      rotate: 360,
    }}
    transition={{ ease: 'linear', duration: 1, repeat: Infinity }}
    style={{ width: size, height: size }}
    data-testid="loading"
  >
    <Flex css={{ color: '$primary' }}>
      <FiLoader size={size} />
    </Flex>
  </motion.div>
)
