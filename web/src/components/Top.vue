<script setup>
import { useRouter } from 'vue-router'
import { useApp } from '@/stores/app'

defineProps({
  title: {
    type: String,
    required: true,
  },
})

const router = useRouter()
const app = useApp()

const logout = () => {
  app.logout()
  router.push({ path: '/' })
}
</script>

<template>
  <div class="wrapper">
    <h1>{{ title }}</h1>

    <nav>
      <ul>
        <li>Vi: {{ app.session.user }}</li>
        <li><RouterLink to="/">Hejmo</RouterLink></li>
        <li><button @click="logout">elsaluti</button></li>
      </ul>
    </nav>
  </div>
</template>

<style scoped>
</style>
