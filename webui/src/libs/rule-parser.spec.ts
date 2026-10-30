import { describe, expect, it } from 'vitest'

import { flattenMatchers, parseRule } from './rule-parser'

describe('rule-parser', () => {
  it('should parse matchers and operators', () => {
    const tokens = parseRule('Host(`a.com`) && PathPrefix(`/api`)')

    expect(tokens).toEqual([
      { type: 'matcher', name: 'Host', args: ['a.com'], negated: false },
      { type: 'operator', value: '&&' },
      { type: 'matcher', name: 'PathPrefix', args: ['/api'], negated: false },
    ])
  })

  it('should flatten matchers of nested groups', () => {
    const tokens = parseRule('(Host(`a.com`) || (Host(`b.com`) && Path(`/x`)))')

    expect(flattenMatchers(tokens).map((m) => m.name)).toEqual(['Host', 'Host', 'Path'])
  })

  it('should fall back to the raw rule when parentheses are unmatched', () => {
    const rule = 'Host(`a.com`'

    expect(parseRule(rule)).toEqual([{ type: 'fallback', raw: rule }])
  })

  it('should fall back to the raw rule when groups are nested too deep', () => {
    const rule = `${'('.repeat(1000)}Host(\`a.com\`)${')'.repeat(1000)}`

    expect(parseRule(rule)).toEqual([{ type: 'fallback', raw: rule }])
  })
})
