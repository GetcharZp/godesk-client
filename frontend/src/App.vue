<template>
  <div class="app-container">
    <!-- 左侧导航栏 -->
    <div class="nav-sidebar">
      <div class="user-info">
        <h2>GoDesk 远程控制</h2>
        <div class="user-detail">
          <p class="username">getcharzp</p>
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
      </nav>
    </div>

    <!-- 主内容区 -->
    <main class="main-content">
      <router-view></router-view>
    </main>
  </div>
</template>

<script setup>

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

.username {
  font-size: 14px;
  font-weight: 500;
}

.main-nav {
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin-bottom: 30px;
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

.main-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}
</style>