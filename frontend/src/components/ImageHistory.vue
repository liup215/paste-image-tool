<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { App } from '../../bindings/github.com/22569/paste-image-tool'

interface ImageInfo {
  path: string
  name: string
  size: number
  width: number
  height: number
  createdAt: string
}

const images = ref<ImageInfo[]>([])
const selectedImage = ref<ImageInfo | null>(null)

onMounted(async () => {
  await loadHistory()
})

const loadHistory = async () => {
  try {
    const history = await App.GetRecentImages()
    if (history) {
      images.value = history.map((item: any) => ({
        path: item.Path,
        name: item.Name,
        size: item.Size,
        width: item.Width,
        height: item.Height,
        createdAt: item.CreatedAt
      }))
    }
  } catch (err) {
    console.error('加载历史记录失败:', err)
  }
}

const formatSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateStr: string): string => {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

const selectImage = (image: ImageInfo) => {
  selectedImage.value = image
}

const copyPath = async (image: ImageInfo) => {
  try {
    await navigator.clipboard.writeText(image.path)
    alert('路径已复制到剪切板')
  } catch (err) {
    console.error('复制失败:', err)
  }
}

const openDirectory = async (image: ImageInfo) => {
  try {
    await App.OpenDirectory()
  } catch (err) {
    console.error('打开目录失败:', err)
  }
}

const deleteImage = async (image: ImageInfo) => {
  // 这里应该调用后端删除图片
  images.value = images.value.filter(img => img.path !== image.path)
  if (selectedImage.value?.path === image.path) {
    selectedImage.value = null
  }
}

const clearHistory = () => {
  images.value = []
  selectedImage.value = null
}
</script>

<template>
  <div class="image-history">
    <div class="history-header">
      <h2>历史记录</h2>
      <button class="btn-clear" @click="clearHistory">清空历史</button>
    </div>
    
    <div class="history-content">
      <div class="image-list">
        <div 
          v-for="image in images" 
          :key="image.path"
          :class="['image-item', { active: selectedImage?.path === image.path }]"
          @click="selectImage(image)"
        >
          <div class="image-preview">
            <img :src="'file://' + image.path" :alt="image.name" />
          </div>
          <div class="image-info">
            <div class="image-name">{{ image.name }}</div>
            <div class="image-meta">
              {{ image.width }}x{{ image.height }} | {{ formatSize(image.size) }}
            </div>
            <div class="image-date">{{ formatDate(image.createdAt) }}</div>
          </div>
        </div>
        
        <div v-if="images.length === 0" class="empty-state">
          <p>暂无历史记录</p>
          <p class="hint">复制图片到剪切板即可自动保存</p>
        </div>
      </div>
      
      <div class="image-detail" v-if="selectedImage">
        <h3>图片详情</h3>
        <div class="detail-preview">
          <img :src="'file://' + selectedImage.path" :alt="selectedImage.name" />
        </div>
        <div class="detail-info">
          <div class="detail-row">
            <span class="label">文件名:</span>
            <span class="value">{{ selectedImage.name }}</span>
          </div>
          <div class="detail-row">
            <span class="label">路径:</span>
            <span class="value path">{{ selectedImage.path }}</span>
          </div>
          <div class="detail-row">
            <span class="label">尺寸:</span>
            <span class="value">{{ selectedImage.width }} x {{ selectedImage.height }}</span>
          </div>
          <div class="detail-row">
            <span class="label">大小:</span>
            <span class="value">{{ formatSize(selectedImage.size) }}</span>
          </div>
          <div class="detail-row">
            <span class="label">创建时间:</span>
            <span class="value">{{ formatDate(selectedImage.createdAt) }}</span>
          </div>
        </div>
        <div class="detail-actions">
          <button @click="copyPath(selectedImage)">复制路径</button>
          <button @click="openDirectory(selectedImage)">打开目录</button>
          <button class="btn-danger" @click="deleteImage(selectedImage)">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.image-history {
  background: #3d3d4a;
  border-radius: 8px;
  padding: 1.5rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  color: #e0e0e0;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
}

.history-header h2 {
  margin: 0;
  color: #fff;
  font-size: 1.1rem;
}

.btn-clear {
  padding: 0.5rem 1rem;
  background: #4d4d5a;
  border: 1px solid #5d5d6a;
  border-radius: 4px;
  color: #e0e0e0;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.85rem;
}

.btn-clear:hover {
  background: #5d3d3d;
  border-color: #ef5350;
  color: #ff9999;
}

.history-content {
  display: grid;
  grid-template-columns: 1fr 280px;
  gap: 1.5rem;
}

.image-list {
  max-height: 450px;
  overflow-y: auto;
}

.image-item {
  display: flex;
  gap: 0.75rem;
  padding: 0.75rem;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
  border: 1px solid transparent;
  background: #2d2d3a;
  margin-bottom: 0.5rem;
}

.image-item:hover {
  background: #4d4d5a;
}

.image-item.active {
  background: #3d4d6a;
  border-color: #667eea;
}

.image-preview {
  width: 60px;
  height: 60px;
  border-radius: 4px;
  overflow: hidden;
  background: #1d1d2a;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.image-preview img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.image-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 0;
}

.image-name {
  font-weight: 500;
  color: #fff;
  font-size: 0.85rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.image-meta {
  font-size: 0.75rem;
  color: #888;
}

.image-date {
  font-size: 0.7rem;
  color: #666;
}

.empty-state {
  text-align: center;
  padding: 3rem;
  color: #888;
}

.empty-state .hint {
  font-size: 0.8rem;
  color: #666;
}

.image-detail {
  border-left: 1px solid #4d4d5a;
  padding-left: 1.5rem;
}

.image-detail h3 {
  margin: 0 0 1rem 0;
  color: #fff;
  font-size: 1rem;
}

.detail-preview {
  width: 100%;
  height: 150px;
  border-radius: 6px;
  overflow: hidden;
  background: #1d1d2a;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.detail-preview img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.detail-info {
  margin-bottom: 1rem;
}

.detail-row {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.4rem;
  font-size: 0.8rem;
}

.detail-row .label {
  color: #888;
  min-width: 70px;
}

.detail-row .value {
  color: #e0e0e0;
  flex: 1;
}

.detail-row .value.path {
  word-break: break-all;
  font-size: 0.7rem;
  color: #888;
}

.detail-actions {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.detail-actions button {
  padding: 0.5rem;
  background: #4d4d5a;
  border: 1px solid #5d5d6a;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  color: #e0e0e0;
  font-size: 0.8rem;
}

.detail-actions button:hover {
  background: #5d5d6a;
}

.detail-actions .btn-danger {
  background: #5d3d3d;
  border-color: #ef5350;
  color: #ff9999;
}

.detail-actions .btn-danger:hover {
  background: #ef5350;
  color: white;
}
</style>