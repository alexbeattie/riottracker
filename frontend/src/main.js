// main.js
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

// Suppress harmless browser extension/Mapbox worker connection errors
window.addEventListener('unhandledrejection', (event) => {
  const errorMsg = event.reason?.message || event.reason?.toString() || '';
  if (errorMsg.includes('Could not establish connection') || 
      errorMsg.includes('Receiving end does not exist')) {
    event.preventDefault(); // Prevent the error from showing in console
    return;
  }
});

const app = createApp(App)
const pinia = createPinia()

// Order matters - initialize pinia before mounting
app.use(pinia)
app.use(router)
app.mount('#app')