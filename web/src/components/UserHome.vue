<script setup>
import { ref } from 'vue'
import { useApp } from '@/stores/app'

const app = useApp()

const courses = ref({ data: null, error: '' })
const lessons = ref(new Map())

// const getLessons = (courseID) => {
//   const lesson = { data: null, error: '' }

//   lessons.set(courseID, l)

//   app.getEntity('/uzantoj/'+ app.session.user_id+'/kursoj/'+courseID+'/lessons')
//     .then(e => {
//       lesson.data = e
//     })
//     .catch(err => {
//       lesson.error = err
//     })
// }

// const getCourses = () => {
//   app.getEntity('/uzantoj/'+ app.session.user_id+'/kursoj')
//     .then(e => {
//       courses.value.data = e
//       for (let course of e) {
//         getLessons(course.id)
//       }
//     })
//     .catch(err => {
//       courses.value.error = err
//     })
// }

// getCourses()
</script>

<template>
  <h2>Miaj kursoj</h2>
  <ul v-for="course in app.getCourses().value">
    <li class="title">{{course}}
      <!-- <ul v-for="lesson in lessons.get(course.id)">
        <li>lesson: {{lesson}}</li>
      </ul> -->
    </li>
  </ul>
</template>
