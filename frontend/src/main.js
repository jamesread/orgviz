import { createApp } from 'vue'

import 'femtocrank/style.css'
import 'picocrank/styles/picocrank-extensions.css'
import 'picocrank/styles/dark-theme.css'
import 'picocrank/vue/composables/useTheme.js'

import './style.css'
import App from './App.vue'
import router from './router.js'

createApp(App).use(router).mount('#app')
