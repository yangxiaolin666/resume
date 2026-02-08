import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore(
  'rec-user',
  () => {
    const username = ref('')
    const userType = ref('')
    const token = ref('')
    const setAll = (u, ut, t) => {
      username.value = u
      userType.value = ut
      token.value = t
    }
    const getAll = () => {
      console.log(username.value, userType.value, token.value)
    }
    return { username, userType, token, setAll, getAll }
  },
  {
    persist: true
  }
)
