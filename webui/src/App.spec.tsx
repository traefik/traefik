import App from './App'

import { render } from 'utils/test'

describe('<App />', () => {
  test('renders without crashing the initial page (dashboard)', () => {
    const { getByText } = render(<App />)
    const text = getByText('Dashboard')
    expect(text).toBeInTheDocument()
  })
})
