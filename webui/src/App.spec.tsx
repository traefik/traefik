import { render } from 'utils/test'

import App from './App'

describe('<App />', () => {
  test('renders without crashing the initial page (dashboard)', () => {
    const { getByText } = render(<App />)
    const text = getByText('Dashboard')
    expect(text).toBeInTheDocument()
  })
})
