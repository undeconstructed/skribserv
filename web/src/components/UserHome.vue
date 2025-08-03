<script setup>
import { ref } from 'vue'
import { useApp } from '@/stores/app'

const app = useApp()

const datumoj = ref({ ento: null, eraro: '' })

app.getEntity('/uzantoj/'+ app.session.user_id+'/kursoj')
  .then(e => {
    datumoj.value.ento = e
  })
  .catch(err => {
    datumoj.value.eraro = err
  })
</script>

<template>
  <h2>Miaj kursoj</h2>
  <p>{{ datumoj.ento }}</p>
  <h2>Miaj hejmtaskoj</h2>
</template>
