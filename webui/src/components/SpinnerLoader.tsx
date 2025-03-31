import { motion } from 'framer-motion'
import { FiLoader } from 'react-icons/fi'

export const SpinnerLoader = () => (
  <motion.div
    animate={{
      rotate: 360,
    }}
    transition={{ ease: 'linear', duration: 1, repeat: Infinity }}
    style={{ width: 24, height: 24 }}
    data-testid="loading"
  >
    <FiLoader color="hsl(222, 67%, 51%)" size={24} />
  </motion.div>
)
