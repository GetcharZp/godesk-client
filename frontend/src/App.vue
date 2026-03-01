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
  background-color: #f5f5f5;
}

.nav-sidebar {
  width: 220px;
  background-color: white;
  padding: 20px;
  display: flex;
  flex-direction: column;
  box-shadow: 1px 0 5px rgba(0, 0, 0, 0.1);
}

.user-info {
  margin-bottom: 30px;
}

.user-info h2 {
  color: #409EFF;
  font-size: 18px;
  margin-bottom: 10px;
}

.user-detail {
  display: flex;
  align-items: center;
  gap: 10px;
}

.main-nav {
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin-bottom: 30px;
  flex: 1;
}

.nav-item {
  padding: 10px 15px;
  color: #666;
  text-decoration: none;
  border-radius: 4px;
  transition: all 0.3s;
}

.nav-item:hover {
  background-color: #ECF5FF;
}

.nav-item.active {
  background-color: #ECF5FF;
  color: #409EFF;
  font-weight: 500;
}

/* 连接状态样式 */
.connection-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 15px;
  margin-top: auto;
  border-top: 1px solid #e8e8e8;
  font-size: 13px;
  color: #666;
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #ff4d4f;
  transition: background-color 0.3s;
}

.status-indicator.connected {
  background-color: #52c41a;
}

.status-text {
  font-size: 13px;
  color: #666;
}

.main-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}
</style>
