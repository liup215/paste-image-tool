<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { App } from '../../bindings/github.com/22569/paste-image-tool'
import { Config } from '../../bindings/github.com/22569/paste-image-tool/internal/config'

const config = ref<Config>({
  saveDirectory: '',
  filenameTemplate: 'paste_{date}_{time}.png',
  imageFormat: 'png',
  jpegQuality: 85,
  hotkey: 'Ctrl+Shift+Insert',
  startMinimized: false,
  pathFormat: 'absolute',
  enableHistory: true,
  maxHistory: 100,
  autoCompress: true,
  maxDimension: 1920
})

const message = ref('')
const messageType = ref('success')

onMounted(async () => {
  try {
    const cfg = await App.GetConfig()
    if (cfg) {
      config.value = cfg
    }
  } catch (err) {
    console.error('加载配置失败:', err)
  }
})

const saveConfig = async () => {
  try {
    // 使用 bindings 中定义的 Config 接口格式（小写字段名）
    const cfg: Config = {
      saveDirectory: config.value.saveDirectory,
      filenameTemplate: config.value.filenameTemplate,
      imageFormat: config.value.imageFormat,
      jpegQuality: config.value.jpegQuality,
      hotkey: config.value.hotkey,
      startMinimized: config.value.startMinimized,
      pathFormat: config.value.pathFormat,
      enableHistory: config.value.enableHistory,
      maxHistory: config.value.maxHistory,
      autoCompress: config.value.autoCompress,
      maxDimension: config.value.maxDimension
    }
    
    await App.UpdateConfig(cfg)
    showMessage('配置已保存', 'success')
  } catch (err) {
    console.error('保存配置错误:', err)
    showMessage('保存失败: ' + err, 'error')
  }
}

const selectDirectory = async () => {
  try {
    const dir = await App.SelectDirectory()
    if (dir) {
      config.value.saveDirectory = dir
    }
  } catch (err) {
    console.error('选择目录失败:', err)
  }
}

const showMessage = (msg: string, type: string) => {
  message.value = msg
  messageType.value = type
  setTimeout(() => {
    message.value = ''
  }, 3000)
}

// 捕获按键设置快捷键
const captureHotkey = (e: KeyboardEvent) => {
  const modifiers: string[] = []
  
  if (e.ctrlKey) modifiers.push('Ctrl')
  if (e.altKey) modifiers.push('Alt')
  if (e.shiftKey) modifiers.push('Shift')
  if (e.metaKey) modifiers.push('Win')
  
  // 获取按键名称
  let key = e.key
  
  // 忽略单独的修饰键
  if (key === 'Control' || key === 'Alt' || key === 'Shift' || key === 'Meta') {
    return
  }
  
  // 转换特殊按键名称
  if (key === ' ') key = 'Space'
  else if (key.length === 1) key = key.toUpperCase()
  
  // 组合快捷键字符串
  if (modifiers.length > 0) {
    config.value.hotkey = modifiers.join('+') + '+' + key
  } else {
    config.value.hotkey = key
  }
}
</script>

<template>
  <div class="config-panel">
    <h2>基本设置</h2>
    
    <div class="form-group">
      <label>保存目录</label>
      <div class="input-row">
        <input 
          v-model="config.saveDirectory" 
          type="text" 
          placeholder="图片保存目录"
        />
        <button @click="selectDirectory">选择</button>
      </div>
    </div>
    
    <div class="form-group">
      <label>文件名模板</label>
      <input 
        v-model="config.filenameTemplate" 
        type="text" 
        placeholder="paste_{date}_{time}.png"
      />
      <small>可用变量: {date}, {time}, {timestamp}</small>
    </div>
    
    <div class="form-group">
      <label>图片格式</label>
      <select v-model="config.imageFormat">
        <option value="png">PNG (无损)</option>
        <option value="jpg">JPEG (压缩)</option>
        <option value="webp">WebP (现代格式)</option>
      </select>
    </div>
    
    <div class="form-group" v-if="config.imageFormat === 'jpg'">
      <label>JPEG 质量 (1-100)</label>
      <input 
        v-model.number="config.jpegQuality" 
        type="number" 
        min="1" 
        max="100"
      />
    </div>
    
    <div class="form-group">
      <label>全局快捷键</label>
      <input 
        v-model="config.hotkey" 
        type="text" 
        placeholder="点击此处按快捷键"
        @keydown.prevent="captureHotkey"
        @keyup.prevent
        readonly
        class="hotkey-input"
      />
      <small>点击输入框，按下想要的快捷键组合</small>
    </div>
    
    <div class="form-group">
      <label>路径格式</label>
      <select v-model="config.pathFormat">
        <option value="absolute">绝对路径</option>
        <option value="relative">相对路径</option>
        <option value="markdown">Markdown</option>
        <option value="html">HTML</option>
        <option value="url">URL</option>
      </select>
    </div>
    
    <h2>高级设置</h2>
    
    <div class="form-group checkbox">
      <label>
        <input v-model="config.autoCompress" type="checkbox" />
        自动压缩大尺寸图片
      </label>
    </div>
    
    <div class="form-group" v-if="config.autoCompress">
      <label>最大尺寸限制 (像素)</label>
      <input 
        v-model.number="config.maxDimension" 
        type="number" 
        min="100" 
        max="4096"
      />
    </div>
    
    <div class="form-group checkbox">
      <label>
        <input v-model="config.enableHistory" type="checkbox" />
        启用历史记录
      </label>
    </div>
    
    <div class="form-group" v-if="config.enableHistory">
      <label>历史记录最大数量</label>
      <input 
        v-model.number="config.maxHistory" 
        type="number" 
        min="10" 
        max="1000"
      />
    </div>
    
    <div class="form-group checkbox">
      <label>
        <input v-model="config.startMinimized" type="checkbox" />
        启动时最小化到托盘
      </label>
    </div>
    
    <div class="actions">
      <button class="btn-primary" @click="saveConfig">保存配置</button>
    </div>
    
    <div v-if="message" :class="['message', messageType]">
      {{ message }}
    </div>
  </div>
</template>

<style scoped>
.config-panel {
  background: #3d3d4a;
  border-radius: 8px;
  padding: 1.5rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  color: #e0e0e0;
}

.config-panel h2 {
  margin: 0 0 1.25rem 0;
  color: #fff;
  font-size: 1.1rem;
  border-bottom: 2px solid #667eea;
  padding-bottom: 0.5rem;
}

.config-panel h2:not(:first-child) {
  margin-top: 1.5rem;
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.4rem;
  color: #b0b0c0;
  font-weight: 500;
  font-size: 0.9rem;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 0.6rem 0.75rem;
  border: 1px solid #5d5d6a;
  border-radius: 4px;
  font-size: 0.9rem;
  transition: border-color 0.2s;
  background: #2d2d3a;
  color: #fff;
}

.form-group input::placeholder {
  color: #888;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #667eea;
}

.form-group select {
  cursor: pointer;
  appearance: menulist;
  -webkit-appearance: menulist;
  background-color: #2d2d3a;
  color: #fff;
  pointer-events: auto;
}

.form-group select option {
  background-color: #2d2d3a;
  color: #fff;
}

.hotkey-input {
  cursor: pointer;
  text-align: center;
  font-weight: 500;
}

.hotkey-input:focus {
  border-color: #667eea;
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.3);
}

.form-group small {
  display: block;
  margin-top: 0.25rem;
  color: #888;
  font-size: 0.8rem;
}

.form-group.checkbox label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  color: #e0e0e0;
}

.form-group.checkbox input {
  width: auto;
  accent-color: #667eea;
}

.input-row {
  display: flex;
  gap: 0.5rem;
}

.input-row input {
  flex: 1;
}

.input-row button {
  padding: 0.6rem 1rem;
  background: #4d4d5a;
  border: 1px solid #5d5d6a;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  color: #e0e0e0;
  font-size: 0.85rem;
}

.input-row button:hover {
  background: #5d5d6a;
}

.actions {
  margin-top: 1.5rem;
  text-align: center;
}

.btn-primary {
  padding: 0.6rem 1.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 0.9rem;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.message {
  margin-top: 1rem;
  padding: 0.6rem;
  border-radius: 4px;
  text-align: center;
  font-size: 0.9rem;
}

.message.success {
  background: #2d5a3d;
  color: #90ee90;
}

.message.error {
  background: #5a3d3d;
  color: #ff9999;
}
</style>