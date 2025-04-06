import { createApp } from 'vue'
import App from './App.vue'

const app = createApp(App)

// Global error handler - avoid using variables that might be undefined
app.config.errorHandler = (err, instance, info) => {
  console.error('Global error:', err)
  console.error('Error info:', info)
  // Log error to console only, let components handle UI error display
}

// Handle uncaught promise rejections
window.addEventListener('unhandledrejection', event => {
  console.error('Unhandled promise rejection:', event.reason)
})

app.mount('#app')