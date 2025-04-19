<script setup>
import { ref } from 'vue'
import { useApp } from '@/stores/app'

const app = useApp()

const mail = ref('')
const pass = ref('')

const data = ref({ working: false, error: ''})

const login = () => {
  if (!mail.value || !pass.value) {
    alert("fill in!")
    return
  }

  data.value.working = true

  app.login(mail, pass)
  .then((u) => {
    data.value.working = false
  })
  .catch(err => {
    data.value.working = false
    data.value.error = err
  })
}
</script>

<template>
  <div class="login">
    <h1>Ensaluti</h1>
    <p v-if="data.error">{{ data.error }}</p>
    <form v-if="!data.working" @submit.prevent="login">
      <p><input type="text" placeholder="retpoŝtadreso" v-model="mail"></p>
      <p><input type="password" placeholder="pasvorto" v-model="pass"></p>
      <p><button type="submit">ensaluti</button></p>
    </form>
    <p v-else>Atendu...</p>
  </div>
</template>

<style>
.login {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
</style>
