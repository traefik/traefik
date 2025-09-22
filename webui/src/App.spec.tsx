import App from './App'

import { render } from 'utils/test'

describe('<App />', () => {
  test('renders without crashing the initial page (dashboard)', () => {
    const { getByTestId } = render(<App />)
    expect(getByTestId('proxy-main-nav')).toBeInTheDocument()
  })
})
