import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore(
  'rec-user',
  () => {
    const userId=ref('')
    const username = ref('')
    const displayName = ref('')
    const userType = ref('')
    const token = ref('')
    const setAll = (id, u, ut, t = '', name = '') => {
      userId.value = id
      username.value = u
      userType.value = ut
      token.value = t
      displayName.value = name || u
    }
    const setDisplayName = (name) => {
      displayName.value = name
    }
    const clear = () => {
      userId.value = ''
      username.value = ''
      displayName.value = ''
      userType.value = ''
      token.value = ''
    }
    const getAll = () => {
      console.log(userId.value,username.value, userType.value, token.value)
    }
    return { userId,username, displayName, userType, token, setAll, setDisplayName, clear, getAll }
  },
  {
    persist: true
  }
)
