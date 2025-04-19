import { ref, toValue } from 'vue'
import { defineStore } from 'pinia'

function getCookie(cname) {
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

function setCookie(cname, cvalue, exdays) {
  const d = new Date()
  d.setTime(d.getTime() + (exdays * 24 * 60 * 60 * 1000))
  let expires = "expires=" + d.toUTCString()
  document.cookie = cname + "=" + cvalue + ";" + expires + ";path=/"
}

export const useApp = defineStore('the', () => {
  const isReady = ref(false)
  const session = ref({
    isLoggedIn: false,
    user: null
  })

  const cache = new Map()

  const check = async () => {
    if (getCookie('Seanco')) {
      const res = await fetch('/api/mi')

      if (res.status != 200) {
        setCookie('Seanco', '', -1)
        session.value.isLoggedIn = false
        session.value.user = null
      } else {
        const json = await res.json()
        session.value.isLoggedIn = true
        session.value.user = json.ento.nomo
      }
    }

    isReady.value = true
  }

  check()

  const onFetchError = (err) => {
    logout()
  }

  const login = async (name, pass) => {
    return new Promise(async (resolve, reject) => {
      const res = await fetch('/api/mi/ensaluti', {
        method: 'POST',
        body: JSON.stringify({
          'retpoŝto': toValue(name),
          'pasvorto': toValue(pass)
        })
      })

      if (res.status == 401) {
        onFetchError()
        reject('not logged in')
        return
      } else if (res.status != 200) {
        reject(`status ${res.status}`)
        return
      }

      const json = await res.json()
      session.value.isLoggedIn = true
      session.value.user = json.ento.nomo

      resolve(json.ento)
    })
  }

  const logout = () => {
    cache.clear()
    setCookie('Seanco', '', -1)
    session.value.isLoggedIn = false
    session.value.user = null
  }

  const getEntity = async (path) => {
    let e = cache.get(path)

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

    cache.set(path, e)

    return e
  }

  const createCourse = async (opts) => {
    const body = JSON.stringify(opts)

    return new Promise(async (resolve, reject) => {
      const res = await fetch('/api/kursoj', {
        method: 'POST',
        body: body
      })

      if (res.status == 401) {
        onFetchError()
        reject('not logged in')
        return
      } else if (res.status != 200) {
        reject(`http status ${res.status}`)
        return
      }

      cache.delete('/kursoj')

      const json = await res.json()

      resolve(json.ento)
    })
  }

  return { isReady, session, login, logout, getEntity, createCourse }
})
