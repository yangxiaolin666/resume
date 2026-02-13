import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore(
  'rec-user',
  () => {
    const userId=ref('')
    const username = ref('')
    const userType = ref('')
    const token = ref('')
    const setAll = (id,u, ut, t) => {
      userId.value=id
      username.value = u
      userType.value = ut
      token.value = t
    }
    const getAll = () => {
      console.log(userId.value,username.value, userType.value, token.value)
    }
    return { userId,username, userType, token, setAll, getAll }
  },
  {
    persist: true
  }
)
