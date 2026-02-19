import { reactive } from 'vue'

const state = reactive({
  show: false,
  message: '',
  type: 'success',
})

let timeoutId = null

export function useToast() {
  const showToast = (message, type = 'success') => {
    if (timeoutId) {
      clearTimeout(timeoutId)
    }

    state.show = true
    state.message = message
    state.type = type

    timeoutId = setTimeout(() => {
      state.show = false
    }, 3000)
  }

  return {
    toast: state,
    showToast,
  }
}
