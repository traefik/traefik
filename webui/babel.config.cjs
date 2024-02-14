/* eslint-disable */

module.exports = api => {
  return {
    presets: [
      [
        '@quasar/babel-preset-app',
        api.caller(caller => caller && caller.target === 'node')
          ? { targets: { node: 'current' } }
          : {}
      ]
    ]
  }
}


