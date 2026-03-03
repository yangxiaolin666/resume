import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useResumeStore = defineStore('resume', () => {
  const savedResume = ref(null)
  
  const saveResume = (resumeData) => {
    savedResume.value = {
      ...resumeData,
      timestamp: new Date().toISOString()
    }
  }
  
  const loadResume = () => {
    return savedResume.value
  }
  
  const hasSavedResume = () => {
    return savedResume.value !== null
  }
  
  const clearResume = () => {
    savedResume.value = null
  }
  
  return {
    savedResume,
    saveResume,
    loadResume,
    hasSavedResume,
    clearResume
  }
}, {
  persist: true
})