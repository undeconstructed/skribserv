<script setup>
import { useRouter } from 'vue-router'
import { ref } from 'vue'
import { useApp } from '@/stores/app'

const router = useRouter()
const app = useApp()

const name = ref('')

const error = ref('')

const create = () => {
  if (!name.value) {
    alert("fill in!")
    return
  }

  app.createCourse({ nomo: name.value })
    .then((k) => {
      router.push({ name: 'course', params: { id: k.id } })
    })
    .catch(e => {
      error.value = e
    })
}
</script>

<template>
  <h1>Krei kurson</h1>

  <p if="error">{{ error }}</p>
  <form @submit.prevent="create">
    <p><input type="text" placeholder="nomo" v-model="name"></p>
    <p><button type="submit">krei</button></p>
  </form>
</template>
