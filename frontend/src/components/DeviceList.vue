<template>
  <div class="device-list">
    <h2 class="page-title">设备列表</h2>

    <div class="search-bar">
      <input type="text" placeholder="搜索设备">
    </div>

    <div class="device-categories">

      <div class="category">
        <h3>我的设备 (3)</h3>
        <div class="device-items">
          <div class="device-item">DESKTOP-459CUIA</div>
          <div class="device-item">DESKTOP-459CUIE</div>
          <div class="device-item">DESKTOP-AGMQ460</div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
</script>

<style scoped>
.device-list {
  max-width: 800px;
  margin: 0 auto;
}

.page-title {
  color: #333;
  margin-bottom: 20px;
  font-size: 20px;
}

.search-bar {
  margin-bottom: 20px;
}

.search-bar input {
  width: 100%;
  padding: 10px 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.device-categories {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
  margin-bottom: 20px;
}

.category {
  margin-bottom: 25px;
}

.category:last-child {
  margin-bottom: 0;
}

.category h3 {
  font-size: 16px;
  color: #333;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.device-items {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.device-item {
  padding: 10px 15px;
  background-color: #f9f9f9;
  border-radius: 4px;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.device-id {
  font-family: monospace;
}

.device-name {
  color: #666;
}

.promotion-section {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
}

.promotion-item {
  flex: 1;
  background-color: white;
  padding: 15px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.promotion-item h4 {
  font-size: 16px;
  color: #333;
  margin-bottom: 5px;
}

.promotion-item p {
  font-size: 14px;
  color: #666;
  margin-bottom: 10px;
}

</style>