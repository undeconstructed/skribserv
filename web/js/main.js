import { PageManagerElement, PageElement, mkel, setCookie, getCookie } from './util.js'

customElements.define('page-manager', class extends PageManagerElement {
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
      admin: false,
      courses: null // [{"id": "k-123", "nomo": "kurso unu", "lessons": []}, {"id": "k-456", "nomo": "kurso du", "lessons": []}],
    }
  }

  connectedCallback() {
    super.connectedCallback()

    this.attachShadow({ mode: 'open' }).appendChild(
      document.getElementById('tmpl-page-manager').content.cloneNode(true)
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
    this.attachShadow({ mode: 'open' }).appendChild(
      document.getElementById('tmpl-main-page').content.cloneNode(true)
    )

    let content = document.querySelector(`[data-name="${this.getAttribute('content')}"]`)
    this.closest('page-manager').fixLinks(content)
    content.setAttribute('slot', 'content')
  }
})

customElements.define('my-courses-page', class extends PageElement {
  #courseList

  connectedCallback() {
    this.courseList = mkel('div', { attrs: { 'slot': 'list' } })
    this.append(this.courseList)

    this.attachShadow({ mode: 'open' }).appendChild(
      document.getElementById('tmpl-my-courses-page').content.cloneNode(true)
    )
  }

  async onShow() {
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
    this.#lessonList = mkel('div', { attrs: { 'slot': 'list' } })
    this.append(this.#lessonList)

    this.attachShadow({ mode: 'open' }).appendChild(
      document.getElementById('tmpl-course-line').content.cloneNode(true)
    )

    let course = await this.manager().getCourse(this.courseID)

    this.setData(course)

    let lessons = await this.manager().getLessons(course.id)

    this.#lessonList.append(
      ...lessons.map(e => mkel('p', { text: `${e.nomo} (${e.id})` }))
    )
  }
})

customElements.define('courses-page', class extends PageElement {
  #courseList

  connectedCallback() {
    this.courseList = mkel('div', { attrs: { 'slot': 'list' } })
    this.append(this.courseList)

    this.attachShadow({ mode: 'open' }).appendChild(
      document.getElementById('tmpl-courses-page').content.cloneNode(true)
    )
  }

  async onShow() {
    this.courseList.replaceChildren()

    let courses = await this.manager().getCourses()

    this.courseList.append(
      ...courses.map(e => mkel('course-line-2', { courseID: e.id }))
    )
  }
})

customElements.define('course-line-2', class extends PageElement {
  async connectedCallback() {
    this.attachShadow({ mode: 'open' }).innerHTML = `
      <p><a href="/$/kursoj/${this.courseID}" data-id="nomo"></a> - <spab data-id="pri"></span></p>
    `

    let course = await this.manager().getCourse(this.courseID)

    this.setData(course)
    this.manager().fixLinks(this.shadowRoot)
  }
})

customElements.define('course-page', class extends PageElement {
  #courseID
  #lessonList

  async connectedCallback() {
    this.#lessonList = mkel('div', { attrs: { 'slot': 'list' } })
    this.append(this.#lessonList)

    this.attachShadow({ mode: 'open' }).appendChild(
      document.getElementById('tmpl-course-page').content.cloneNode(true)
    )
  }

  async onShow() {
    this.#courseID = this.getAttribute('course-id')

    let course = await this.manager().getCourse(this.#courseID)

    this.setData(course)

    let lessons = await this.manager().getLessons(course.id)

    this.#lessonList.append(
      ...lessons.map(e => mkel('p', { text: `${e.nomo} (${e.id})` }))
    )
  }
})

customElements.define('data-block', class extends PageElement {
})
