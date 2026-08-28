export type MatcherToken = {
  type: 'matcher'
  name: string
  args: string[]
  negated: boolean
}

export type OperatorToken = {
  type: 'operator'
  value: '&&' | '||'
}

export type GroupToken = {
  type: 'group'
  tokens: Token[]
  negated: boolean
}

export type FallbackToken = {
  type: 'fallback'
  raw: string
}

export type Token = MatcherToken | OperatorToken | GroupToken | FallbackToken

export function parseRule(rule: string): Token[] {
  if (!rule) return []
  try {
    return parseTokens(rule.trim())
  } catch {
    return [{ type: 'fallback', raw: rule }]
  }
}

function parseTokens(input: string): Token[] {
  const tokens: Token[] = []
  let i = 0
  let negated = false

  while (i < input.length) {
    if (/\s/.test(input[i])) {
      i++
      continue
    }

    if (input[i] === '!') {
      negated = true
      i++
      continue
    }

    if (input.slice(i, i + 2) === '&&') {
      tokens.push({ type: 'operator', value: '&&' })
      i += 2
      continue
    }

    if (input.slice(i, i + 2) === '||') {
      tokens.push({ type: 'operator', value: '||' })
      i += 2
      continue
    }

    if (input[i] === '(') {
      const closeIdx = findMatchingParen(input, i)
      const inner = input.slice(i + 1, closeIdx)
      tokens.push({ type: 'group', tokens: parseTokens(inner), negated })
      negated = false
      i = closeIdx + 1
      continue
    }

    const matcherMatch = /^([A-Za-z]+)\(/.exec(input.slice(i))
    if (matcherMatch) {
      const name = matcherMatch[1]
      const argsStart = i + name.length
      const argsEnd = findMatchingParen(input, argsStart)
      const argsStr = input.slice(argsStart + 1, argsEnd)
      tokens.push({ type: 'matcher', name, args: extractArgs(argsStr), negated })
      negated = false
      i = argsEnd + 1
      continue
    }

    i++
  }

  return tokens
}

function findMatchingParen(input: string, openIdx: number): number {
  let depth = 0
  let inBacktick = false
  for (let i = openIdx; i < input.length; i++) {
    const ch = input[i]
    if (ch === '`') {
      inBacktick = !inBacktick
      continue
    }
    if (inBacktick) continue
    if (ch === '(') depth++
    else if (ch === ')') {
      depth--
      if (depth === 0) return i
    }
  }
  throw new Error('Unmatched parenthesis')
}

function extractArgs(argsStr: string): string[] {
  const args: string[] = []
  let i = 0
  while (i < argsStr.length) {
    if (/[\s,]/.test(argsStr[i])) {
      i++
      continue
    }
    if (argsStr[i] === '`') {
      const end = argsStr.indexOf('`', i + 1)
      if (end === -1) break
      args.push(argsStr.slice(i + 1, end))
      i = end + 1
      continue
    }
    i++
  }
  return args
}

export function flattenMatchers(tokens: Token[]): MatcherToken[] {
  const result: MatcherToken[] = []
  for (const t of tokens) {
    if (t.type === 'matcher') result.push(t)
    else if (t.type === 'group') result.push(...flattenMatchers(t.tokens))
  }
  return result
}
