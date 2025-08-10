
export function setCookie(cname, cvalue, exdays) {
  const d = new Date()
  d.setTime(d.getTime() + (exdays * 24 * 60 * 60 * 1000))
  let expires = "expires=" + d.toUTCString()
  document.cookie = cname + "=" + cvalue + ";" + expires + ";path=/"
}

export function getCookie(cname) {
  let name = cname + '='
  for (let c of document.cookie.split(';')) {
    while (c.charAt(0) == ' ') {
      c = c.substring(1)
    }
    if (c.indexOf(name) == 0) {
      return c.substring(name.length)
    }
  }
  return ''
}

export function* zip(...arrays){
  const maxLength = arrays.reduce((max, curIterable) => curIterable.length > max ? curIterable.length: max, 0);
  for (let i = 0; i < maxLength; i++) {
    yield arrays.map(array => array[i]);
  }
}

export function join(a, s, f) {
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
      case 'attrs':
        let attrs = opts[opt]
        for (let a in attrs) {
          e.setAttribute(a, attrs[a])
        }
        break
      default:
        e[opt] = opts[opt]
    }
  }
  return e
}

export function defer(delay, func, args) {
  return setTimeout(function () {
    return func(...args)
  }, delay)
}

export function hook(src, event, options, func, args) {
  return src.addEventListener(event, function (e) {
    return func(e, ...args)
  }, options)
}

export function switchy(arg, opts) {
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

export function select(parent, selector) {
  return parent.querySelector(selector)
}

// PageManagerElement is a base for page management.
export class PageManagerElement extends HTMLElement {
  #title
  #paths
  #slot = 'content'

  constructor(opts) {
    super()
    if (opts) {
      this.#slot = opts.slot || 'content'
    }
  }

  connectedCallback() {
    this.#title = document.title
    this.#paths = new Router(Array.from(this.querySelectorAll('[path]')).map(e => ({ path: e.getAttribute('path'), element: e })))

    window.addEventListener('popstate', e => {
      this.showPage(this.parseHarsh(), true)
    })
  }

  fixLinks(top) {
    if (!top) return

    for (let link of top.querySelectorAll('a[href*="/$/"]')) {
      let m = link.href.match(/\/\$(\/.*)/)
      if (m) {
        let to = m[1]
        if (!to.startsWith('/')) {
          to = '/' + to
        }
        link.href = '/#' + to
        link.addEventListener('click', e => {
          e.preventDefault()
          this.showPage(to)
        })
      }
    }
  }

  parseHarsh() {
    let hash = URL.parse(window.location.href).hash
    if (hash.startsWith('#/')) {
      hash = hash.substring(1)
    } else {
      hash = '/'
    }
    return hash
  }

  route(path) {
    return this.#paths.route(path)
  }

  showPage(path, replace) {
    let c = this.route(path)

    if (path != '/') {
      path = '/#' + path
    }

    if (replace) history.replaceState({}, '', path)
    else history.pushState({}, '', path)

    let to = c.element

    for (let k in c.path) {
      to.setAttribute(k, c.path[k])
    }

    this.swapPage(to)
  }

  swapPage(to, notitle) {
    if (this.open) {
      this.open.removeAttribute('slot')
      this.open.onHide && this.open.onHide()
    }

    if (typeof to == 'string') {
      to = this.querySelector(to)
    }

    if (!to) {
      to = this.querySelector('#error')
      to.setAttribute('page-title', '404')
      to.querySelector('.msg').textContent = '404'
    }

    if (!notitle) {
      if (this.route('/').element === to) {
        document.title = this.#title
      } else {
        let title = (to.pageTitle && to.pageTitle()) || to.getAttribute('page-title') || to.tagName
        document.title = this.#title + ' - ' + title
      }
    }

    this.fixLinks(to)
    this.fixLinks(to.shadowRoot)

    this.open = to
    this.open.onShow && this.open.onShow()
    this.open.setAttribute('slot', this.#slot)
  }
}

// Router has some logic for handling internal links.
export class Router {
  #data

  constructor(rules) {
    this.#data = rules.map(this.parseRule)
  }

  parseRule(rule) {
    if (rule.path.endsWith('/')) {
      rule.path = rule.path.slice(0, -1)
    }

    let path = rule.path.split('/').map(e => {
      if (e == '') return { root: true }

      let m = e.match(/\{(.*)\}/)
      if (m) {
        return { id: m[1] }
      }

      return { lit: e }
    })
    return {
      path: path,
      element: rule.element,
    }
  }

  match(url, rule) {
    if (url.endsWith('/')) {
      url = url.slice(0, -1)
    }

    let path = url.split('/')
    if (path.length != rule.path.length) {
      return null
    }

    let m = {}

    for (let [a, b] of [...zip(path, rule.path)]) {
      if (a == '' && b.root) {
        continue
      }
      if (a == b.lit) {
        continue
      }
      if (b.id) {
        m[b.id] = a
        continue
      }
      m = null
      break
    }

    return m
  }

  route(url) {
    for (let rule of this.#data) {
      let m = this.match(url, rule)
      if (m) {
        return {
          element: rule.element,
          path: m
        }
      }
    }

    return null
  }
}

// PageElement helps a custom element to be a managed page.
export class PageElement extends HTMLElement {
  manager() {
    return this.closest('page-manager')
  }

  inSlot() {
    return !!this.getAttribute('slot')
  }

  setData(o) {
    for (let e of this.shadowRoot.querySelectorAll('[data-id]')) {
      let did = e.getAttribute('data-id')
      e.textContent = o[did]
    }
  }
}

export function main(func) {
  document.addEventListener('DOMContentLoaded', function() {
    func({
      window,
      document,
      localStorage
    })
  })
}
