import { get } from 'dot-prop'

class Helps {
  // Getters
  // ------------------------------------------------------------------------

  // Public
  // ------------------------------------------------------------------------

  // Static
  // ------------------------------------------------------------------------

  static get (obj, prop, def = undefined) {
    return get(obj, prop, def)
  }

  static hasIn (obj, prop) {
    return Helps.get(obj, prop) !== undefined && Helps.get(obj, prop) !== null
  }

  static toggleBodyClass (addRemoveClass, className) {
    const el = document.body

    if (addRemoveClass === 'addClass') {
      el.classList.add(className)
    } else {
      el.classList.remove(className)
    }
  }

  static getName (obj, val) {
    let name = ''
    for (let i = 0; i < obj.length; i += 1) {
      if (obj[i].value === val || obj[i].iso2 === val) {
        name = obj[i].name
      }
    }
    return name
  }

  static removeEmptyObjects (objects) {
    const obj = {}
    Object.entries(objects).map(item => {
      if (item[1] !== '') {
        obj[item[0]] = item[1]
      }
    })
    return obj
  }

  // Helps -> Numbers
  // ------------------------------------------------------------------------

  static getPercent (value, total) {
    return (value * 100) / total
  }

  // Helps -> Array
  // ------------------------------------------------------------------------

  // Add or remove values
  static toggleArray (array, value) {
    if (array.includes(value)) {
      array.splice(array.indexOf(value), 1)
    } else {
      array.push(value)
    }
  }

  // Helps -> Strings
  // ------------------------------------------------------------------------

  // Basename
  static basename (path, suffix) {
    let b = path
    const lastChar = b.charAt(b.length - 1)

    if (lastChar === '/' || lastChar === '\\') {
      b = b.slice(0, -1)
    }

    // eslint-disable-next-line no-useless-escape
    b = b.replace(/^.*[\/\\]/g, '')

    if (typeof suffix === 'string' && b.substr(b.length - suffix.length) === suffix) {
      b = b.substr(0, b.length - suffix.length)
    }

    return b
  }

  // Slug
  static slug (str) {
    str = str.replace(/^\s+|\s+$/g, '') // trim
    str = str.toLowerCase()

    // remove accents, swap ñ for n, etc
    const from = 'ãàáäâẽèéëêìíïîõòóöôùúüûñç·/_,:;'
    const to = 'aaaaaeeeeeiiiiooooouuuunc------'
    for (let i = 0, l = from.length; i < l; i += 1) {
      str = str.replace(new RegExp(from.charAt(i), 'g'), to.charAt(i))
    }

    str = str.replace(/[^a-z0-9 -]/g, '') // remove invalid chars
      .replace(/\s+/g, '-') // collapse whitespace and replace by -
      .replace(/-+/g, '-') // collapse dashes

    return str
  }

  // Capitalize first letter
  static capFirstLetter (string) {
    return string.charAt(0).toUpperCase() + string.slice(1)
  }

  // Repeat
  static repeat (string, times) {
    return new Array(times + 1).join(string)
  }

  // Get Attribute
  static getAttribute (string, key) {
    const _key = `${key}="`
    const start = string.indexOf(_key) + _key.length
    const end = string.indexOf('"', start + 1)
    return string.substring(start, end)
  }

  // Private
  // ------------------------------------------------------------------------
}

export default Helps
