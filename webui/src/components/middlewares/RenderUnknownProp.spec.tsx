import { RenderUnknownProp } from './RenderUnknownProp'

import { renderWithProviders } from 'utils/test'

describe('<RenderUnknownProp />', () => {
  it('renders a string correctly', () => {
    const { container } = renderWithProviders(<RenderUnknownProp name="StringPropName" prop="string prop value" />)

    expect(container.querySelector('div > span')?.innerHTML).toContain('StringPropName')
    expect(container.querySelector('div > div')?.innerHTML).toContain('string prop value')
  })

  it('renders a number correctly', () => {
    const { container } = renderWithProviders(<RenderUnknownProp name="NumberPropName" prop={123123} />)

    expect(container.querySelector('div > span')?.innerHTML).toContain('NumberPropName')
    expect(container.querySelector('div > div')?.innerHTML).toContain('123123')
  })

  it('renders false correctly', () => {
    const { container } = renderWithProviders(<RenderUnknownProp name="BooleanPropName" prop={false} />)

    expect(container.querySelector('div > span')?.innerHTML).toContain('BooleanPropName')
    expect(container.querySelector('div > div')?.innerHTML).toContain('enabled-false')
    expect(container.querySelector('div > div')?.innerHTML).toContain('False')
  })

  it('renders boolean true correctly', () => {
    const { container } = renderWithProviders(<RenderUnknownProp name="BooleanPropName" prop={true} />)

    expect(container.querySelector('div > span')?.innerHTML).toContain('BooleanPropName')
    expect(container.querySelector('div > div')?.innerHTML).toContain('enabled-true')
    expect(container.querySelector('div > div')?.innerHTML).toContain('True')
  })

  it('renders boolean false correctly', () => {
    const { container } = renderWithProviders(<RenderUnknownProp name="BooleanPropName" prop={false} />)

    expect(container.querySelector('div > span')?.innerHTML).toContain('BooleanPropName')
    expect(container.querySelector('div > div')?.innerHTML).toContain('enabled-false')
    expect(container.querySelector('div > div')?.innerHTML).toContain('False')
  })

  it('renders string `true` correctly', () => {
    const { container } = renderWithProviders(<RenderUnknownProp name="StringBoolPropName" prop="true" />)

    expect(container.querySelector('div > span')?.innerHTML).toContain('StringBoolPropName')
    expect(container.querySelector('div > div')?.innerHTML).toContain('enabled-true')
    expect(container.querySelector('div > div')?.innerHTML).toContain('True')
  })

  it('renders string `false` correctly', () => {
    const { container } = renderWithProviders(<RenderUnknownProp name="StringBoolPropName" prop="false" />)

    expect(container.querySelector('div > span')?.innerHTML).toContain('StringBoolPropName')
    expect(container.querySelector('div > div')?.innerHTML).toContain('enabled-false')
    expect(container.querySelector('div > div')?.innerHTML).toContain('False')
  })

  it('renders empty object correctly', () => {
    const { container } = renderWithProviders(<RenderUnknownProp name="EmptyObjectPropName" prop={{}} />)

    expect(container.querySelector('div > span')?.innerHTML).toContain('EmptyObjectPropName')
    expect(container.querySelector('div > div')?.innerHTML).toContain('enabled-true')
    expect(container.querySelector('div > div')?.innerHTML).toContain('True')
  })

  it('renders list of strings correctly', () => {
    const { container } = renderWithProviders(
      <RenderUnknownProp name="StringListPropName" prop={['string1', 'string2', 'string3']} />,
    )

    expect(container.querySelector('div > span')?.innerHTML).toContain('StringListPropName')
    expect(container.querySelector('div > div')?.innerHTML).toContain('string1')
    expect(container.querySelector('div > div')?.innerHTML).toContain('string2')
    expect(container.querySelector('div > div')?.innerHTML).toContain('string3')
  })

  it('renders list of objects correctly', () => {
    const { container } = renderWithProviders(
      <RenderUnknownProp
        name="ObjectListPropName"
        prop={[{ array: [] }, { otherObject: {} }, { word: 'test' }, { number: 123 }, { boolean: false, or: true }]}
      />,
    )

    expect(container.querySelector('div > span')?.innerHTML).toContain('ObjectListPropName')
    expect(container.querySelector('div > div')?.innerHTML).toContain('{"array":[]}')
    expect(container.querySelector('div > div')?.innerHTML).toContain('{"otherObject":{}}')
    expect(container.querySelector('div > div')?.innerHTML).toContain('{"word":"test"}')
    expect(container.querySelector('div > div')?.innerHTML).toContain('{"number":123}')
    expect(container.querySelector('div > div')?.innerHTML).toContain('{"boolean":false,"or":true}')
  })

  it('renders recursive objects correctly', () => {
    const { container } = renderWithProviders(
      <RenderUnknownProp
        name="RecursiveObjectPropName"
        prop={{
          parentProperty: {
            childProperty: {
              valueProperty1: 'test',
              valueProperty2: ['item1', 'item2', 'item3'],
            },
          },
        }}
      />,
    )

    expect(container.querySelector('div:first-child > span')?.innerHTML).toContain(
      'RecursiveObjectPropName &gt; parent Property &gt; child Property &gt; value Property1',
    )
    expect(container.querySelector('div:first-child > div')?.innerHTML).toContain('test')
    expect(container.querySelector('div:first-child ~ div > span')?.innerHTML).toContain(
      'RecursiveObjectPropName &gt; parent Property &gt; child Property &gt; value Property2',
    )
    expect(container.querySelector('div:first-child ~ div > div')?.innerHTML).toContain('item1')
    expect(container.querySelector('div:first-child ~ div > div')?.innerHTML).toContain('item2')
    expect(container.querySelector('div:first-child ~ div > div')?.innerHTML).toContain('item3')
  })

  it('renders recursive objects removing title prefix correctly', () => {
    const { container } = renderWithProviders(
      <RenderUnknownProp
        name="RecursiveObjectPropName"
        removeTitlePrefix="RecursiveObjectPropName"
        prop={{
          parentProperty: {
            childProperty: {
              valueProperty1: 'test',
              valueProperty2: ['item1', 'item2', 'item3'],
            },
          },
        }}
      />,
    )

    expect(container.querySelector('div:first-child > span')?.innerHTML).toContain(
      'parent Property &gt; child Property &gt; value Property1',
    )
    expect(container.querySelector('div:first-child > div')?.innerHTML).toContain('test')
    expect(container.querySelector('div:first-child ~ div > span')?.innerHTML).toContain(
      'parent Property &gt; child Property &gt; value Property2',
    )
    expect(container.querySelector('div:first-child ~ div > div')?.innerHTML).toContain('item1')
    expect(container.querySelector('div:first-child ~ div > div')?.innerHTML).toContain('item2')
    expect(container.querySelector('div:first-child ~ div > div')?.innerHTML).toContain('item3')
  })

  it(`renders should not remove prefix if there's no child`, () => {
    const { container } = renderWithProviders(
      <RenderUnknownProp
        name="RecursiveObjectPropName"
        removeTitlePrefix="RecursiveObjectPropName"
        prop="DummyValue"
      />,
    )

    expect(container.querySelector('div > span')?.innerHTML).toContain('RecursiveObjectPropName')
    expect(container.querySelector('div > div')?.innerHTML).toContain('DummyValue')
  })
})
