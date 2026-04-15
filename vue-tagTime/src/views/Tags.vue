<template>
  <div class="container">
    <AppHeader />

    <div class="page-header">
      <h1 class="page-title">标签管理</h1>
      <button class="btn btn-primary" @click="showCreateModal = true">+ 新建标签</button>
    </div>

    <div class="tags-grid">
      <div v-if="tags.length === 0" class="empty-state">
        <div class="empty-state-icon">🏷️</div>
        <div class="empty-state-text">暂无标签，点击右上角创建第一个标签吧</div>
      </div>

      <div v-for="tag in tags" :key="tag.id" class="tag-item">
        <div class="tag-color" :style="{ backgroundColor: tag.color }"></div>
        <div class="tag-info">
          <h3>{{ tag.name }}</h3>
          <p>创建于 {{ formatDate(tag.created_at) }}</p>
        </div>
        <div class="tag-actions">
          <button @click="editTag(tag)">编辑</button>
          <button @click="deleteTag(tag.id)">删除</button>
        </div>
      </div>
    </div>

    <!-- 创建/编辑标签弹窗 -->
    <div v-if="showCreateModal" class="modal" @click.self="showCreateModal = false">
      <div class="modal-content">
        <h2>{{ editingTag ? '编辑标签' : '新建标签' }}</h2>
        <div class="form-group">
          <label>标签名称</label>
          <input v-model="tagForm.name" type="text" placeholder="输入标签名称" />
        </div>
        <div class="form-group">
          <label>标签颜色</label>
          <div class="color-picker">
            <div 
              v-for="color in presetColors" 
              :key="color"
              class="color-option"
              :class="{ active: tagForm.color === color }"
              :style="{ backgroundColor: color }"
              @click="tagForm.color = color"
            ></div>
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showCreateModal = false">取消</button>
          <button class="btn btn-primary" @click="saveTag">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { tagAPI, type Tag } from '../api/tag'
import AppHeader from '../components/AppHeader.vue'
import { toast, confirm } from '../utils/message'

const tags = ref<Tag[]>([])
const showCreateModal = ref(false)
const editingTag = ref<Tag | null>(null)

const tagForm = ref({
  name: '',
  color: '#4a90e2'
})

const presetColors = [
  '#4a90e2', '#7ed321', '#f5a623', '#bd10e0', 
  '#50e3c2', '#ff6b6b', '#9013fe', '#f8e71c'
]

const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const loadTags = async () => {
  try {
    const res: any = await tagAPI.getTags()
    tags.value = res.tags || []
  } catch (err) {
    console.error('加载标签失败', err)
    toast.error('加载标签失败')
  }
}

const editTag = (tag: Tag) => {
  editingTag.value = tag
  tagForm.value = {
    name: tag.name,
    color: tag.color
  }
  showCreateModal.value = true
}

const saveTag = async () => {
  if (!tagForm.value.name) {
    toast.warning('请输入标签名称')
    return
  }

  try {
    if (editingTag.value) {
      await tagAPI.updateTag(editingTag.value.id, tagForm.value)
      toast.success('标签已更新')
    } else {
      await tagAPI.createTag(tagForm.value)
      toast.success('标签已创建')
    }
    showCreateModal.value = false
    editingTag.value = null
    tagForm.value = { name: '', color: '#4a90e2' }
    loadTags()
  } catch (err) {
    console.error('保存标签失败', err)
    toast.error('保存标签失败')
  }
}

const deleteTag = async (id: number) => {
  const confirmed = await confirm('确定删除这个标签吗？', '删除标签', { type: 'danger' })
  if (!confirmed) return
  
  try {
    await tagAPI.deleteTag(id)
    toast.success('标签已删除')
    loadTags()
  } catch (err) {
    console.error('删除标签失败', err)
    toast.error('删除标签失败')
  }
}

onMounted(() => {
  loadTags()
})
</script>

<style scoped>
.tags-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.tag-item {
  background-color: #fafafa;
  border: 1px solid #e5e5e5;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 15px;
  transition: all 0.2s;
}

.tag-item:hover {
  border-color: #999;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

.tag-color {
  width: 50px;
  height: 50px;
  border-radius: 8px;
  flex-shrink: 0;
}

.tag-info {
  flex: 1;
}

.tag-info h3 {
  font-size: 18px;
  margin-bottom: 5px;
}

.tag-info p {
  font-size: 13px;
  color: #999;
}

.tag-actions {
  display: flex;
  gap: 10px;
}

.tag-actions button {
  background: none;
  border: none;
  color: #999;
  cursor: pointer;
  font-size: 13px;
  padding: 4px 8px;
}

.tag-actions button:hover {
  color: #333;
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  padding: 30px;
  border-radius: 8px;
  width: 90%;
  max-width: 500px;
}

.modal-content h2 {
  margin-bottom: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
}

.form-group input {
  width: 100%;
  padding: 12px;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  font-size: 14px;
}

.color-picker {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
}

.color-option {
  width: 100%;
  aspect-ratio: 1;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: 3px solid transparent;
}

.color-option:hover {
  transform: scale(1.1);
}

.color-option.active {
  border-color: #333;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
