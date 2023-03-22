<template>
  <!-- we must remove all references to this -->

  <div class="container">
    <div class="row">
      <div class="col">
        <h1 class="mt-3">Books</h1>
      </div>

      <hr />
      <div class="filters text-center">
        <span
          class="filter me-1"
          v-bind:class="{ active: currentFilter === 0 }"
          v-on:click="setFilter(0)"
          >ALL</span
        >
        <span
          class="filter me-1"
          v-bind:class="{ active: currentFilter === 7 }"
          v-on:click="setFilter(7)"
          >CLASSIC</span
        >
        <span
          class="filter me-1"
          v-bind:class="{ active: currentFilter === 2 }"
          v-on:click="setFilter(2)"
          >FANTASY</span
        >
        <span
          class="filter me-1"
          v-bind:class="{ active: currentFilter === 6 }"
          v-on:click="setFilter(6)"
          >HORROR</span
        >
        <span
          class="filter me-1"
          v-bind:class="{ active: currentFilter === 4 }"
          v-on:click="setFilter(4)"
          >THRILLER</span
        >
        <span
          class="filter me-1"
          v-bind:class="{ active: currentFilter === 1 }"
          v-on:click="setFilter(1)"
          >SCIENCE FICTION</span
        >
      </div>
      <hr />
      <div v-if="ready">
        <div class="card">
          <TransitionGroup
            class="p-3 d-flex flex-wrap"
            tag="div"
            appear
            name="books">
            <div v-for="b in books" v-bind:key="b.id">
              <div
                class="card me-2 ms-1 mb-3"
                style="width: 10rem"
                v-if="
                  b.genre_ids.includes(currentFilter) || currentFilter === 0
                ">
                <router-link :to="`/books/${b.slug}`">
                  <img
                    :src="`${imgPath}/covers/${b.slug}.jpg`"
                    class="card-img-top"
                    :alt="`cover for ${b.title}`" />
                </router-link>
                <div class="card-body text-center">
                  <h6 class="card-title">{{ b.title }}</h6>
                  <span class="book-author">{{ b.author.author_name }}</span
                  ><br />
                  <small
                    class="text-muted book-genre"
                    v-for="(g, index) in b.genres"
                    v-bind:key="g.id">
                    <em class="me-1"
                      >{{ g.genre_name
                      }}<template v-if="index !== b.genres.length - 1"
                        >,</template
                      ></em
                    >
                  </small>
                </div>
              </div>
            </div>
          </TransitionGroup>
        </div>
      </div>
      <p v-else>Loading...</p>
    </div>
  </div>
</template>

<script lang="js">
import {ref, onMounted} from 'vue'

export default {
    name: 'BooksComposition',
    emits: ['error'],
    props: {},

    setup(props, ctx) {
        // set up state for this component
        let ready = ref(false);
        let currentFilter = ref(0);
        const imgPath = ref(process.env.VUE_APP_IMAGE_URL)
        let books = ref({})

        // use onMounted lifeCycle hook to get books
        onMounted(() => {
            console.log('using books with composition api')
            fetch(process.env.VUE_APP_API_URL + "/books")
            .then((response) => response.json())
            .then((response) => {
                if (response.error) {
                    ctx.$emit('error', response.message)
                } else {
                    books.value = response.data.books
                    ready.value = true
                }
            })
            .catch(error => {
                ctx.emit('error', error)
            })
        })

        function setFilter(filter) {
            currentFilter.value = filter;
        }

        // return data and functions
        return {
            currentFilter,
            imgPath,
            books,
            setFilter,
            ready,
        }
    }
}
</script>

<style scoped>
.filters {
  height: 2.5em;
}

.filter {
  padding: 6px 6px;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.35;
  border: 1px solid silver;
}

.filter.active {
  background: lightgreen;
}

.filter:hover {
  background: lightgray;
}

.book-author,
.book-genre {
  font-size: 0.8em;
}

/* transition styles */
.books-move {
  transition: all 500ms ease-in-out 50ms;
}

.books-enter-active {
  transition: all 500ms ease-in-out;
}

.books-leave-active {
  transition: all 500ms ease-in;
}

.books-enter,
.books-leave-to {
  opacity: 0;
}
</style>
