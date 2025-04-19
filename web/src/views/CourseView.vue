<script setup>
import { ref } from 'vue'
import { useApp } from '@/stores/app'
import { useRoute } from 'vue-router'

const app = useApp()
const route = useRoute()

const datumoj = ref({ ento: null, lernantoj: null, eraro: '' })

app.getEntity('/kursoj/' + route.params.id)
  .then(e => {
    datumoj.value.ento = e
  })
  .catch(err => {
    datumoj.value.eraro = err
  })

app.getEntity('/kursoj/' + route.params.id + "/lernantoj")
  .then(e => {
    datumoj.value.lernantoj = e
  })
  .catch(err => {
    datumoj.value.eraro = err
  })
</script>

<template>

  <template v-if="datumoj.ento">
    <h1>Kurso: {{ datumoj.ento.nomo }}</h1>
    <p>De: {{ datumoj.ento.posedanto.nomo }}</p>
    <template v-if="datumoj.lernantoj">
      <p>Lernantoj:</p>
      <ul>
        <li v-for="l of datumoj.lernantoj">{{ l.id }}{{ l.nomo }}</li>
      </ul>
    </template>
  </template>
  <template v-else-if="datumoj.eraro">
    <p>{{ datumoj.eraro }}</p>
  </template>
  <template v-else>
    <p>Ŝarganta...</p>
  </template>

</template>
