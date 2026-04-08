<template>
  <div class="app-container">
    <!-- 左侧导航栏 -->
    <div class="nav-sidebar">
      <div class="user-info">
        <h2>GoDesk 远程控制</h2>
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

      <!-- 连接状态 -->
      <div class="connection-status">
        <div class="status-indicator" :class="{ connected: isConnected }"></div>
        <span class="status-text">{{ isConnected ? '已连接' : '未连接' }}</span>
      </div>
    </div>

    <!-- 主内容区 -->
    <main class="main-content">
      <router-view></router-view>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import UserInfo from "./components/UserInfo.vue"
import { getConnectionStatus } from "./api/sys.js"

// 连接状态
const isConnected = ref(false)
let statusTimer = null

// 获取连接状态
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
  // 立即获取一次状态
  fetchConnectionStatus()
  // 定时获取连接状态（每3秒）
  statusTimer = setInterval(fetchConnectionStatus, 3000)
})

onUnmounted(() => {
  // 清除定时器
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
  background: linear-gradient(135deg, #0a0e27 0%, #151b3d 100%);
}

.nav-sidebar {
  width: 220px;
  background: linear-gradient(180deg, #151b3d 0%, #0f1428 100%);
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  box-shadow: 2px 0 10px rgba(0, 0, 0, 0.3);
  border-right: 1px solid #2d3561;
}

.user-info {
  margin-bottom: 32px;
}

.user-info h2 {
  color: #00d4ff;
  font-size: 18px;
  margin-bottom: 16px;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.5);
  letter-spacing: 0.5px;
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
  color: #94a3b8;
  text-decoration: none;
  border-radius: 8px;
  transition: all 0.3s ease;
  border: 1px solid transparent;
  font-size: 14px;
}

.nav-item:hover {
  background: rgba(0, 212, 255, 0.08);
  color: #00d4ff;
  border-color: #2d3561;
}

.nav-item.active {
  background: rgba(0, 212, 255, 0.15);
  color: #00d4ff;
  border-color: #00d4ff;
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.2);
  font-weight: 500;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px;
  margin-top: auto;
  border-top: 1px solid #2d3561;
  font-size: 13px;
  background: rgba(21, 27, 61, 0.3);
  border-radius: 8px;
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background-color: #ef4444;
  box-shadow: 0 0 10px rgba(239, 68, 68, 0.5);
  transition: all 0.3s ease;
}

.status-indicator.connected {
  background-color: #10b981;
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.5);
}

.status-text {
  font-size: 13px;
  color: #94a3b8;
}

.main-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  background: #0a0e27;
}

.main-content::-webkit-scrollbar {
  width: 8px;
}

.main-content::-webkit-scrollbar-track {
  background: #0a0e27;
}

.main-content::-webkit-scrollbar-thumb {
  background: #2d3561;
  border-radius: 4px;
}

.main-content::-webkit-scrollbar-thumb:hover {
  background: #3d4571;
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
