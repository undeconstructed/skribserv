import { ref, watchEffect, toValue } from 'vue'

export function useFetch(url) {
  const data = ref(null)
  const error = ref(null)

  const fetchData = (url) => {
    // reset state before fetching
    data.value = null
    error.value = null

    fetch(url)
      .then((res) => {
        if (res.status == 200)
          return res.json()
        else
          throw `http status ${res.status}`
      })
      .then((json) => (data.value = json))
      .catch((err) => (error.value = err))
  }

  watchEffect(() => {
    fetchData(toValue(url))
  })

  return { data, error }
}
