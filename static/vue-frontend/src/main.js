import { createApp } from 'vue'
import App from './App.vue'

// Import global styles
import './assets/analytics.css'

// Performance optimization - create app with performance options
const app = createApp(App)

// Global error handler - avoid using variables that might be undefined
app.config.errorHandler = (err, _instance, info) => {
  console.error('Global error:', err)
  console.error('Error info:', info)
  // Log error to console only, let components handle UI error display
}

// Handle uncaught promise rejections
window.addEventListener('unhandledrejection', event => {
  console.error('Unhandled promise rejection:', event.reason)
})

// Performance monitoring
if (process.env.NODE_ENV === 'development') {
  app.config.performance = true
}

// Register global directives
app.directive('click-outside', {
  mounted(el, binding) {
    el._clickOutside = (event) => {
      if (!(el === event.target || el.contains(event.target))) {
        binding.value(event)
      }
    }
    document.body.addEventListener('click', el._clickOutside)
  },
  unmounted(el) {
    document.body.removeEventListener('click', el._clickOutside)
  }
})

// Mount the app
app.mount('#app')