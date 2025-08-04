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
    <h1><RouterLink to="/">{{ title }}</RouterLink></h1>

    <nav>
      <ul>
        <li>Vi: {{ app.session.user }}</li>
        <li><button @click="logout">elsaluti</button></li>
      </ul>
    </nav>
  </div>
</template>

<style scoped>
.wrapper {
  display: flex;
  flex-flow: row nowrap;
  align-items: center;
}

.wrapper h1 {
  flex-grow: 2;
}

nav ul li {
  display: inline;
}

nav ul li+li {
  margin-left: .5em;
}
nav ul li+li::before {
  content: "|";
  margin-right: .5em;
}
</style>
