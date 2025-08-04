import { main, mkel, setCookie, getCookie } from './util.js'

class PageElement extends HTMLElement {
  manager() {
    return this.closest('page-manager')
  }
}

customElements.define('page-manager', class extends HTMLElement {
  #title
  #paths
  #state
  #cache

  constructor() {
    super()
    this.resetState()
  }

  resetState() {
    this.#cache = new Map()
    this.#state = {
      ready: false,
      user: null,
      admin: null,
      courses: null // [{"id": "k-123", "nomo": "kurso unu"}, {"id": "k-456", "nomo": "kurso du"}],
    }
  }

  async getCourses() {
    let res = await this.getEntity('/uzantoj/' + this.#state.user.id + '/kursoj')

    this.#state.courses = res

    return this.#state.courses
  }

  async getCourse(id) {
    let courses = await this.getCourses()

    return courses.find(e => e.id == id)
  }

  async getLessons(courseID) {
    let course = await this.getCourse(courseID)

    let res = await this.getEntity('/kursoj/' + course.id + '/eroj')

    course.lessons = res

    return course.lessons
  }

  connectedCallback() {
    this.#title = document.title

    this.#paths = new Map()
    for (let c of this.querySelectorAll('[path]')) {
      this.#paths.set(c.getAttribute('path'), c)
    }

    window.addEventListener('popstate', e => {
      console.log(e)
      this.showPage(this.parseHarsh(), true)
    })

    const template = document.getElementById('tmpl-page-manager')
    const templateContent = template.content

    this.attachShadow({ mode: 'open' }).appendChild(
      templateContent.cloneNode(true)
    )

    this.shadowRoot.querySelector('.user button').addEventListener('click', e => {
      e.preventDefault()
      this.logout()
    })

    this.swapPage('loading-page', true)

    this.setup()
  }

  async setup() {
    if (getCookie('Seanco')) {
      const res = await fetch('/api/mi')

      if (res.status != 200) {
        setCookie('Seanco', '', -1)
        this.swapPage('login-page', true)
      } else {
        const json = await res.json()
        this.onUserReady(json.ento)
      }
    } else {
      this.swapPage('login-page', true)
    }

    this.#state.ready = true
  }

  onUserReady(user) {
    this.#state.user = user
    this.#state.admin = user.admina

    this.shadowRoot.querySelector('.username').textContent = this.#state.user.nomo
    this.shadowRoot.querySelector('div').classList.add('logged-in')

    this.showPage(this.parseHarsh(), true)
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
        if (to != '/') {
          link.href = '/#' + to
        }
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
    return this.#paths.get(path)
  }
  
  showPage(path, replace) {
    let c = this.route(path)

    if (path != '/') {
      path = '/#' + path
    }

    if (replace) history.replaceState({}, '', path)
    else history.pushState({}, '', path)

    this.swapPage(c)
  }

  swapPage(to, notitle) {
    if (this.open) {
      this.open.removeAttribute('slot')
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
      if (this.route('/') === to) {
        document.title = this.#title
      } else {
        let title = (to.pageTitle && to.pageTitle()) || to.getAttribute('page-title') || to.tagName
        document.title = this.#title + ' - ' + title
      }
    }

    this.open = to
    this.open.setAttribute('slot', 'content')

    this.fixLinks(this.open)
    this.fixLinks(this.open.shadowRoot)
  }

  async login(name, pass) {
    return new Promise(async (resolve, reject) => {
      const res = await fetch('/api/mi/ensaluti', {
        method: 'POST',
        body: JSON.stringify({
          'retpoŝto': name,
          'pasvorto': pass
        })
      })

      if (res.status == 401) {
        reject('not logged in')
        return
      } else if (res.status != 200) {
        reject(`status ${res.status}`)
        return
      }

      const json = await res.json()

      this.onUserReady(json.ento)
      resolve(json.ento)
    })
  }

  logout() {
    this.swapPage('loading-page', true)

    fetch('/api/mi/elsaluti', {
      method: 'POST'
    })

    this.shadowRoot.querySelector('div').classList.remove('logged-in')
    this.shadowRoot.querySelector('.username').textContent = ''

    this.resetState()
    setCookie('Seanco', '', -1)

    history.pushState({}, '', '/')
    this.swapPage('login-page', true)
  }

  async getEntity(path) {
    let e = this.#cache.get(path)

    if (e) return e

    e = new Promise(async (resolve, reject) => {
      const res = await fetch('/api' + path)

      if (res.status == 401) {
        onFetchError()
        reject('not logged in')
        return
      } else if (res.status != 200) {
        reject(`http status ${res.status}`)
        return
      }

      const json = await res.json()
      console.log('e', path, json)

      resolve(json.ento)
    })

    this.#cache.set(path, e)

    return e
  }
})

customElements.define('loading-page', class extends PageElement {
  connectedCallback() {
    console.log('connected', this)
  }
})

customElements.define('login-page', class extends HTMLElement {
  connectedCallback() {
    const form = this.querySelector('form')
    form.addEventListener('submit', async e => {
      e.preventDefault()

      const u = form.querySelector('[name=user]').value
      const p = form.querySelector('[name=pass]').value

      this.classList.add('working')

      const manager = this.closest('page-manager')

      try {
        const user = await manager.login(u, p)
      } catch (e) {
        this.querySelector('.error').textContent = e
      }

      this.classList.remove('working')
    })
  }
})

customElements.define('main-page', class extends HTMLElement {
  connectedCallback() {
    const template = document.getElementById('tmpl-main-page')
    const templateContent = template.content

    this.attachShadow({ mode: 'open' }).appendChild(
      templateContent.cloneNode(true)
    )

    let content = document.querySelector(`[data-name="${this.getAttribute('content')}"]`)
    this.closest('page-manager').fixLinks(content)
    content.setAttribute('slot', 'content')
  }
})

customElements.define('my-courses', class extends PageElement {
  static observedAttributes = ['slot']

  #courseList

  connectedCallback() {
    this.courseList = mkel('div', { attrs: { 'slot': 'list' } })
    this.append(this.courseList)

    const template = document.getElementById('tmpl-my-courses')
    const templateContent = template.content

    this.attachShadow({ mode: 'open' }).appendChild(
      templateContent.cloneNode(true)
    )
  }

  async attributeChangedCallback(name, oldValue, newValue) {
    if (name != 'slot' || !newValue) return

    this.courseList.replaceChildren()

    let courses = await this.manager().getCourses()

    this.courseList.append(
      ...courses.map(e => mkel('course-line', { courseID: e.id }))
    )
  }
})

customElements.define('course-line', class extends PageElement {
  #lessonList

  async connectedCallback() {
    this.lessonList = mkel('div', { attrs: { 'slot': 'list' } })
    this.append(this.lessonList)

    const template = document.getElementById('tmpl-course-line')
    const templateContent = template.content

    this.attachShadow({ mode: 'open' }).appendChild(
      templateContent.cloneNode(true)
    )

    this.lessonList.replaceChildren()

    let course = await this.manager().getCourse(this.courseID)

    for (let e of this.shadowRoot.querySelectorAll('[data-id]')) {
      let did = e.getAttribute('data-id')
      e.textContent = course[did]
    }

    let lessons = await this.manager().getLessons(course.id)

    this.lessonList.append(
      ...lessons.map(e => mkel('p', { text: `${e.nomo} (${e.id})` }))
    )
  }
})

customElements.define('course-page', class extends PageElement {
})

customElements.define('data-block', class extends PageElement {
})
