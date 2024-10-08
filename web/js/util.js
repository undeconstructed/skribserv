
export function join (a, s, f) {
  f = f || (e => e.toString())
  s = (s === null ? ',' : s)
  let size = a.length || a.size
  if (size == 0) {
    return ''
  } else if (size == 1) {
    let i = a[Symbol.iterator]()
    return f(i.next().value)
  } else {
    let i = a[Symbol.iterator]()
    let out = f(i.next().value)
    for (let e = i.next(); !e.done; e = i.next()) {
      let n = f(e.value)
      if (n !== null) {
        out += s + f(e.value)
      }
    }
    return out
  }
}

export function mkel(tag, opts) {
  opts = opts || {}
  let e = document.createElement(tag)
  for (let opt in opts) {
    switch (opt) {
      case 'classes':
        e.classList.add(...opts.classes)
        break
      case 'text':
        e.textContent = opts.text
        break
      default:
        e[opt] = opts[opt]
    }
  }
  return e
}

export function defer (delay, func, args) {
  return setTimeout(function () {
    return func(...args)
  }, delay)
}

export function hook (src, event, options, func, args) {
  return src.addEventListener(event, function (e) {
    return func(e, ...args)
  }, options)
}

export function switchy (arg, opts) {
  let o = opts[arg]
  if (!o) {
    o = opts['default']
  }
  if (!o) {
    throw new Error('switchy: unhandled ' + arg)
  }
  if (typeof o === 'function') {
    return o()
  }
  return o
}

export function select (parent, selector) {
  return parent.querySelector(selector)
}

export function main (func) {
  document.addEventListener('DOMContentLoaded', function() {
    func({
      window,
      document,
      localStorage
    })
  })
}
