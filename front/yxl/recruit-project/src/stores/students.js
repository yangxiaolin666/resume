import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useFilterStore = defineStore('filter', () => {
  const checkDegree = ref('')
  const checkMajor = ref('')
  const checkGender = ref('')
  const checkLanguage = ref([])
  const checkJob = ref([])
  const checkSpecial = ref([])
  const checkCity = ref([])
  
  const clearAllFilters = () => {
    checkDegree.value = ''
    checkMajor.value = ''
    checkGender.value = ''
    checkLanguage.value = []
    checkJob.value = []
    checkSpecial.value = []
    checkCity.value = []
  }
  
  return {
    checkDegree,
    checkMajor,
    checkGender,
    checkLanguage,
    checkJob,
    checkSpecial,
    checkCity,
    clearAllFilters
  }
},{
  persist:true
})
