import { ref, watch } from 'vue'
import { WindowSetDarkTheme, WindowSetLightTheme, WindowSetBackgroundColour } from '../../wailsjs/runtime/runtime'

const THEME_KEY = 'godesk-theme'

const isDark = ref(true)

const initTheme = () => {
  const savedTheme = localStorage.getItem(THEME_KEY)
  if (savedTheme !== null) {
    isDark.value = savedTheme === 'dark'
  }
  applyTheme()
}

const applyTheme = () => {
  if (isDark.value) {
    document.documentElement.setAttribute('data-theme', 'dark')
    WindowSetDarkTheme()
    WindowSetBackgroundColour(10, 14, 39, 255)
  } else {
    document.documentElement.setAttribute('data-theme', 'light')
    WindowSetLightTheme()
    WindowSetBackgroundColour(248, 250, 252, 255)
  }
}

const toggleTheme = () => {
  isDark.value = !isDark.value
  localStorage.setItem(THEME_KEY, isDark.value ? 'dark' : 'light')
  applyTheme()
}

watch(isDark, () => {
  applyTheme()
})

export function useTheme() {
  return {
    isDark,
    initTheme,
    toggleTheme
  }
}
