<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ConfigPanel from './components/ConfigPanel.vue'
import ImageHistory from './components/ImageHistory.vue'

const activeTab = ref('config')
</script>

<template>
  <div class="app">
    <header class="app-header">
      <h1>Paste Image Tool</h1>
      <p>剪切板图片粘贴工具</p>
    </header>
    
    <nav class="app-nav">
      <button 
        :class="{ active: activeTab === 'config' }" 
        @click="activeTab = 'config'"
      >
        配置
      </button>
      <button 
        :class="{ active: activeTab === 'history' }" 
        @click="activeTab = 'history'"
      >
        历史记录
      </button>
    </nav>
    
    <main class="app-main">
      <ConfigPanel v-if="activeTab === 'config'" />
      <ImageHistory v-if="activeTab === 'history'" />
    </main>
  </div>
</template>

<style scoped>
.app {
  height: 100vh;
  background: #2d2d3a;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.app-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 1rem 2rem;
  text-align: center;
  flex-shrink: 0;
}

.app-header h1 {
  margin: 0 0 0.25rem 0;
  font-size: 1.5rem;
}

.app-header p {
  margin: 0;
  opacity: 0.9;
  font-size: 0.875rem;
}

.app-nav {
  display: flex;
  justify-content: center;
  gap: 1rem;
  padding: 0.75rem;
  background: #3d3d4a;
  border-bottom: 1px solid #4d4d5a;
  flex-shrink: 0;
}

.app-nav button {
  padding: 0.5rem 1.5rem;
  border: none;
  background: transparent;
  color: #b0b0c0;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s;
  font-size: 0.9rem;
}

.app-nav button:hover {
  background: #4d4d5a;
  color: white;
}

.app-nav button.active {
  background: #667eea;
  color: white;
}

.app-main {
  flex: 1;
  padding: 1.5rem;
  overflow: auto;
  display: flex;
  justify-content: center;
}

.app-main > * {
  width: 100%;
  max-width: 700px;
}
</style>
