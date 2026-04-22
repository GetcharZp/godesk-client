<template>
  <div class="app-container">
    <div class="nav-sidebar">
      <div class="user-info">
        <div class="title-row">
          <h2>GoDesk 远程控制</h2>
          <button class="theme-toggle-btn" @click="toggleTheme" :title="isDark ? '切换到亮色模式' : '切换到暗色模式'">
            <svg v-if="isDark" class="theme-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="12" cy="12" r="5" stroke="currentColor" stroke-width="2"/>
              <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            </svg>
            <svg v-else class="theme-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </button>
        </div>
        <div class="user-detail">
          <UserInfo/>
        </div>
      </div>

      <nav class="main-nav">
        <router-link
            to="/remote-control"
            class="nav-item"
            :class="{ active: $route.path === '/remote-control' }"
        >
          远程控制
        </router-link>
        <router-link
            to="/device-list"
            class="nav-item"
            :class="{ active: $route.path === '/device-list' }"
        >
          设备列表
        </router-link>
        <router-link
            to="/system-settings"
            class="nav-item"
            :class="{ active: $route.path === '/system-settings' }"
        >
          系统设置
        </router-link>
      </nav>

      <div class="connection-status">
        <div class="status-indicator" :class="{ connected: isConnected }"></div>
        <span class="status-text">{{ isConnected ? '已连接' : '未连接' }}</span>
      </div>
    </div>

    <main class="main-content">
      <router-view></router-view>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import UserInfo from "./components/UserInfo.vue"
import { getConnectionStatus } from "./api/sys.js"
import { useTheme } from "./composables/useTheme.js"

const { isDark, initTheme, toggleTheme } = useTheme()

const isConnected = ref(false)
let statusTimer = null

const fetchConnectionStatus = async () => {
  try {
    const res = await getConnectionStatus()
    if (res && res.code === 200) {
      isConnected.value = res.data === true
    }
  } catch (error) {
    isConnected.value = false
  }
}

onMounted(() => {
  initTheme()
  fetchConnectionStatus()
  statusTimer = setInterval(fetchConnectionStatus, 3000)
})

onUnmounted(() => {
  if (statusTimer) {
    clearInterval(statusTimer)
  }
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
  font-family: 'PingFang SC', 'Microsoft YaHei', sans-serif;
}

.app-container {
  display: flex;
  min-height: 100vh;
  background: var(--bg-primary);
}

.nav-sidebar {
  width: 220px;
  background: var(--bg-secondary);
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-glow);
  border-right: 1px solid var(--border-primary);
}

.user-info {
  margin-bottom: 32px;
}

.title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.user-info h2 {
  color: var(--text-accent);
  font-size: 16px;
  text-shadow: 0 0 10px var(--accent-primary-glow);
  letter-spacing: 0.5px;
}

.theme-toggle-btn {
  background: transparent;
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  color: var(--text-secondary);
}

.theme-toggle-btn:hover {
  border-color: var(--border-active);
  color: var(--text-accent);
  box-shadow: 0 0 10px var(--accent-primary-glow);
}

.theme-icon {
  width: 20px;
  height: 20px;
}

.user-detail {
  display: flex;
  align-items: center;
  gap: 10px;
}

.main-nav {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 32px;
  flex: 1;
}

.nav-item {
  padding: 12px 16px;
  color: var(--text-secondary);
  text-decoration: none;
  border-radius: 8px;
  transition: all 0.3s ease;
  border: 1px solid transparent;
  font-size: 14px;
}

.nav-item:hover {
  background: var(--bg-card-hover);
  color: var(--text-accent);
  border-color: var(--border-primary);
}

.nav-item.active {
  background: var(--bg-item-hover);
  color: var(--text-accent);
  border-color: var(--border-active);
  box-shadow: 0 0 15px var(--accent-primary-glow);
  font-weight: 500;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px;
  margin-top: auto;
  border-top: 1px solid var(--border-primary);
  font-size: 13px;
  background: var(--bg-tertiary);
  border-radius: 8px;
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background-color: var(--danger);
  box-shadow: 0 0 10px var(--danger-glow);
  transition: all 0.3s ease;
}

.status-indicator.connected {
  background-color: var(--success);
  box-shadow: 0 0 10px var(--success-glow);
}

.status-text {
  font-size: 13px;
  color: var(--text-secondary);
}

.main-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  background: var(--bg-tertiary);
}

.main-content::-webkit-scrollbar {
  width: 8px;
}

.main-content::-webkit-scrollbar-track {
  background: var(--scrollbar-track);
}

.main-content::-webkit-scrollbar-thumb {
  background: var(--scrollbar-thumb);
  border-radius: 4px;
}

.main-content::-webkit-scrollbar-thumb:hover {
  background: var(--scrollbar-thumb-hover);
}

body.remote-control-fullscreen .nav-sidebar {
  display: none;
}

body.remote-control-fullscreen .main-content {
  padding: 0;
}

body.remote-control-fullscreen .app-container {
  display: block;
}
</style>
