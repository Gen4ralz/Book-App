<template>
  <div class="container">
    <div class="row">
      <div class="col">
        <h1 class="mt-4">All Users</h1>
      </div>
      <hr />
      <table v-if="this.ready" class="table table-compact table-striped">
        <thead>
          <tr>
            <th>User</th>
            <th>Email</th>
            <th>Active</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in this.users" v-bind:key="u.id">
            <td>
              <router-link :to="`/admin/users/${u.id}`"
                >{{ u.first_name }} {{ u.last_name }}</router-link
              >
            </td>
            <td>{{ u.email }}</td>
            <td v-if="u.active === 1">
              <span class="badge bg-success">Active</span>
            </td>
            <td v-else>
              <span class="badge bg-danger">Inactive</span>
            </td>
            <td v-if="u.token.id > 0">
              <a href="javascript:void(0);">
                <span class="badge bg-success" @click="logUserOut(u.id)"
                  >Logged in</span
                >
              </a>
            </td>
            <td v-else>
              <span class="badge bg-danger">Not logged in</span>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-else>Loading...</p>
    </div>
  </div>
</template>

<script en="js">
import Security from './security'
import notie from 'notie'
import { store } from './store'

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

export default {
  data() {
    return {
      users: {},
      ready: false,
      store,
    }
  },
  beforeMount() {
    Security.requireToken()

    fetch(
      process.env.VUE_APP_API_URL + '/admin/users',
      Security.requestOptions('')
    )
      .then((response) => response.json())
      .then((response) => {
        if (response.error) {
          notie.alert({
            type: 'error',
            text: response.message,
          })
        } else {
          sleep(3000).then(() => {
            this.users = response.data.users
            this.ready = true
          })
        }
      })
      .catch((error) => {
        notie.alert({
          type: 'error',
          text: error,
        })
      })
  },
  methods: {
    logUserOut(id) {
      if (id !== store.user.id) {
        notie.confirm({
          text: 'Are you sure you want to log this user out?',
          submitText: 'Log Out!',
          submitCallback: () => {
            console.log('Would log out user id', id)
            fetch(
              process.env.VUE_APP_API_URL + '/admin/logout/' + id,
              Security.requestOptions('')
            )
              .then((response) => response.json())
              .then((data) => {
                if (data.error) {
                  this.$emit('error', data.message)
                } else {
                  this.$emit('success', data.message)
                  this.$emit('forceUpdate')
                }
              })
          },
        })
      } else {
        this.$emit('error', "You can't log yourself out!")
      }
    },
  },
}
</script>
