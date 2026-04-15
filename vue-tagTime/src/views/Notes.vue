<template>
  <div class="container">
    <AppHeader />

    <div class="page-header">
      <h1 class="page-title">{{ showTrash ? '回收站' : '全部笔记' }}</h1>
      <div class="header-actions">
        <button 
          v-if="!showTrash" 
          class="btn btn-secondary" 
          @click="toggleTrash"
          title="回收站"
        >
          🗑️ 回收站
        </button>
        <button 
          v-if="showTrash" 
          class="btn btn-secondary" 
          @click="toggleTrash"
        >
          ← 返回笔记
        </button>
        <button 
          v-if="showTrash && notes.length > 0" 
          class="btn btn-danger" 
          @click="confirmEmptyTrash"
        >
          清空回收站
        </button>
      </div>
    </div>

    <div class="main-content">
      <div class="notes-list">
        <!-- 聊天框输入区域 -->
        <div class="chat-input-container">
          <div class="chat-input-box">
            <div 
              class="chat-textarea"
              contenteditable="true"
              @input="handleInput"
              @keydown="handleKeydown"
              @keyup="handleKeyup"
              @click="handleTextareaClick"
              @paste="handlePaste"
              ref="chatTextarea"
              :data-placeholder="quickNoteContent ? '' : '现在的想法是...'"
            ></div>
            
            <div class="chat-toolbar">
              <div class="toolbar-left">
                <button class="toolbar-btn" @click="insertTag" title="添加标签">
                  #
                </button>
                <button class="toolbar-btn" @click="triggerImageUpload" title="上传图片">
                  📷
                </button>
                <button class="toolbar-btn" title="文本格式">
                  Aa
                </button>
                <button class="toolbar-btn" title="有序列表">
                  ≡
                </button>
                <button class="toolbar-btn" title="无序列表">
                  ≡
                </button>
                <button class="toolbar-btn" @click="insertTask" title="关联任务">
                  @
                </button>
              </div>
              
              <div class="toolbar-right">
                <span v-if="getTextContent().trim()" class="char-count">{{ getTextContent().length }}</span>
                <button 
                  class="send-btn" 
                  :class="{ active: getTextContent().trim() }"
                  :disabled="!getTextContent().trim() || isSending"
                  @click="sendQuickNote"
                >
                  ▶
                </button>
              </div>
            </div>

            <!-- 标签下拉选择器 -->
            <div v-if="showTagDropdown" class="mention-dropdown" :style="dropdownStyle">
              <div class="mention-list">
                <div 
                  v-for="tag in filteredTags" 
                  :key="tag.id" 
                  class="mention-item"
                  @mousedown.prevent="selectTag(tag)"
                >
                  <span :style="{ color: tag.color }">{{ tag.name }}</span>
                </div>
                <div v-if="filteredTags.length === 0" class="mention-empty">
                  没有匹配的标签
                </div>
              </div>
            </div>

            <!-- 任务下拉选择器 -->
            <div v-if="showTaskDropdown" class="mention-dropdown" :style="dropdownStyle">
              <div class="mention-list">
                <div 
                  v-for="task in filteredTasks" 
                  :key="task.id" 
                  class="mention-item"
                  @mousedown.prevent="selectTask(task)"
                >
                  <span>{{ task.name }}</span>
                </div>
                <div v-if="filteredTasks.length === 0" class="mention-empty">
                  没有可用的任务
                </div>
              </div>
            </div>

            <!-- 快速上传的图片预览 -->
            <div v-if="quickUploadedImages.length > 0" class="quick-images-preview">
              <div v-for="(img, index) in quickUploadedImages" :key="index" class="quick-image-item">
                <img :src="img" alt="预览图" />
                <button class="remove-quick-image" @click="removeQuickImage(index)">×</button>
              </div>
            </div>
          </div>
        </div>
        <div v-if="filteredNotes.length === 0" class="empty-state">
          <div class="empty-state-icon">{{ showTrash ? '🗑️' : '📝' }}</div>
          <div class="empty-state-text">
            {{ showTrash ? '回收站是空的' : '暂无笔记，点击右上角创建第一条笔记吧' }}
          </div>
        </div>

        <div v-for="note in filteredNotes" :key="note.id" class="note-card" :class="{ 'trash-note': showTrash }" @click="viewNoteDetail(note)">
          <div class="note-header">
            <span class="note-date">{{ formatDate(note.created_at) }}</span>
            <div class="note-actions">
              <button v-if="!showTrash" @click.stop="editNote(note)">编辑</button>
              <button v-if="!showTrash" @click.stop="deleteNote(note.id)">删除</button>
              <button v-if="showTrash" @click.stop="restoreNote(note.id)" class="btn-restore">恢复</button>
              <button v-if="showTrash" @click.stop="permanentDeleteNote(note.id)" class="btn-permanent-delete">永久删除</button>
            </div>
          </div>
          <div class="note-content" v-html="highlightNoteContent(note.content)"></div>
          <div v-if="note.images" class="note-images">
            <img 
              v-for="(img, idx) in parseImages(note.images)" 
              :key="idx"
              :src="img" 
              alt="笔记图片"
              class="note-image"
              @click.stop="previewImage(img)"
            />
          </div>
          <div v-if="showTrash && note.deleted_at" class="trash-info">
            删除于 {{ formatDate(note.deleted_at) }}
          </div>
        </div>
      </div>

      <aside class="sidebar">
        <div class="sidebar-section">
          <div class="sidebar-title">月度记录</div>
          <div class="calendar-header">
            <button class="calendar-nav" @click="changeMonth(-1)">‹</button>
            <span class="calendar-month">{{ currentYear }}年{{ currentMonth }}月</span>
            <button class="calendar-nav" @click="changeMonth(1)">›</button>
          </div>
          <div class="calendar-grid">
            <div v-for="day in ['日','一','二','三','四','五','六']" :key="day" class="calendar-day-header">
              {{ day }}
            </div>
            <div 
              v-for="(day, index) in calendarDays" 
              :key="index"
              class="calendar-day"
              :class="{ 
                empty: day === 0, 
                'no-record': day > 0 && !hasRecord(day),
                'has-record': day > 0 && hasRecord(day)
              }"
            >
              {{ day || '' }}
            </div>
          </div>
        </div>

        <div class="sidebar-section">
          <div class="sidebar-title">本月统计</div>
          <div style="font-size: 14px; color: #666; line-height: 2;">
            <div>总笔记数：<span style="color: #333; font-weight: 600;">{{ notes.length }}</span></div>
            <div>记录天数：<span style="color: #333; font-weight: 600;">{{ recordDays }}</span></div>
          </div>
        </div>

        <div class="sidebar-section">
          <div class="tag-filter-header">
            <span class="sidebar-title">全部标签</span>
            <button class="sort-btn" @click="toggleSortOrder">排序</button>
          </div>
          <div class="tag-filter-list">
            <div 
              v-for="tag in sortedTags" 
              :key="tag.id" 
              class="tag-filter-item"
              :class="{ active: selectedTagId === tag.id }"
            >
              <div class="tag-info" @click="selectTagFilter(tag.id)">
                <span class="tag-hash">#</span>
                <span class="tag-name">{{ tag.name }}</span>
                <span class="tag-count">{{ getTagNoteCount(tag.id) }}</span>
              </div>
              <button class="tag-delete-btn" @click.stop="showDeleteTagModal(tag)" title="删除标签">×</button>
            </div>
            <div 
              v-if="selectedTagId !== null"
              class="tag-filter-item clear-filter"
              @click="clearTagFilter"
            >
              <span class="tag-name">清除筛选</span>
            </div>
          </div>
        </div>
      </aside>
    </div>

    <!-- 笔记详情弹窗 -->
    <div v-if="showDetailModal" class="modal" @click.self="closeDetailModal">
      <div class="modal-content detail-modal">
        <div class="detail-header">
          <h2>笔记详情</h2>
          <button class="close-btn" @click="closeDetailModal">×</button>
        </div>
        <div class="detail-body">
          <div class="detail-date">{{ formatDate(viewingNote?.created_at || '') }}</div>
          <div class="detail-content" v-html="highlightNoteContent(viewingNote?.content || '')"></div>
          <div v-if="viewingNote?.images" class="detail-images">
            <img 
              v-for="(img, idx) in parseImages(viewingNote.images)" 
              :key="idx"
              :src="img" 
              alt="笔记图片"
              class="detail-image"
              @click="previewImage(img)"
            />
          </div>
        </div>
        <div class="detail-actions">
          <button class="btn btn-secondary" @click="editNote(viewingNote!)">编辑</button>
          <button class="btn btn-danger" @click="deleteNote(viewingNote!.id)">删除</button>
        </div>
      </div>
    </div>

    <!-- 图片预览弹窗 -->
    <div v-if="showImagePreview" class="modal image-preview-modal" @click="closeImagePreview">
      <div class="image-preview-content">
        <button class="close-btn" @click="closeImagePreview">×</button>
        <img :src="previewImageUrl" alt="预览图片" />
      </div>
    </div>

    <!-- 创建/编辑笔记弹窗 -->
    <div v-if="showCreateModal" class="modal" @click.self="closeModal">
      <div class="modal-content">
        <h2>{{ editingNote ? '编辑笔记' : '新建笔记' }}</h2>
        <textarea 
          v-model="noteForm.content" 
          placeholder="写点什么..."
          rows="6"
        ></textarea>
        
        <div class="form-group">
          <label>添加图片</label>
          <div class="image-upload-area">
            <input 
              type="file" 
              id="imageInput"
              accept="image/*"
              multiple
              @change="handleImageUpload"
              style="display: none"
            />
            <label for="imageInput" class="upload-btn">
              <span v-if="!isUploading">📷 选择图片</span>
              <span v-else>上传中...</span>
            </label>
            <div v-if="uploadedImages.length > 0" class="uploaded-images">
              <div v-for="(img, index) in uploadedImages" :key="index" class="image-preview">
                <img :src="img" alt="预览图" />
                <button class="remove-image" @click="removeImage(index)">×</button>
              </div>
            </div>
          </div>
        </div>

        <div class="form-group">
          <label>选择标签</label>
          <div class="tag-select">
            <label v-for="tag in tags" :key="tag.id" class="tag-checkbox">
              <input 
                type="checkbox" 
                :value="tag.id"
                v-model="noteForm.tag_ids"
              />
              <span :style="{ color: tag.color }">{{ tag.name }}</span>
            </label>
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="closeModal">取消</button>
          <button class="btn btn-primary" @click="saveNote" :disabled="isUploading">保存</button>
        </div>
      </div>
    </div>

    <!-- 删除标签确认弹窗 -->
    <div v-if="showDeleteTagConfirm" class="modal" @click.self="showDeleteTagConfirm = false">
      <div class="modal-content delete-tag-modal">
        <h2>删除标签</h2>
        <div class="delete-tag-info">
          <p class="tag-name-display">
            <span class="tag-hash">#</span>{{ deletingTag?.name }}
          </p>
          <p class="tag-note-count">该标签关联了 <strong>{{ getTagNoteCount(deletingTag?.id || 0) }}</strong> 条笔记</p>
        </div>
        <div class="delete-options">
          <label class="delete-option">
            <input type="radio" v-model="deleteTagMode" value="tag-only" />
            <div class="option-content">
              <div class="option-title">仅删除标签</div>
              <div class="option-desc">从所有笔记中移除 #{{ deletingTag?.name }} 标签引用，保留笔记内容</div>
            </div>
          </label>
          <label class="delete-option">
            <input type="radio" v-model="deleteTagMode" value="tag-and-notes" />
            <div class="option-content">
              <div class="option-title">删除标签和笔记</div>
              <div class="option-desc warning">⚠️ 将永久删除该标签及其关联的所有笔记，此操作不可恢复</div>
            </div>
          </label>
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showDeleteTagConfirm = false">取消</button>
          <button class="btn btn-danger" @click="confirmDeleteTag">确认删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { noteAPI, type Note } from '../api/note'
import { tagAPI, type Tag } from '../api/tag'
import { taskAPI } from '../api/task'
import AppHeader from '../components/AppHeader.vue'
import { toast, confirm } from '../utils/message'

// 定义组件名称以支持 keep-alive
defineOptions({
  name: 'Notes'
})

const notes = ref<Note[]>([])
const tags = ref<Tag[]>([])
const selectedTagId = ref<number | null>(null)
const showCreateModal = ref(false)
const showDetailModal = ref(false)
const showImagePreview = ref(false)
const editingNote = ref<Note | null>(null)
const viewingNote = ref<Note | null>(null)
const previewImageUrl = ref('')
const calendarData = ref<Record<string, number>>({})
const showTrash = ref(false) // 是否显示回收站

const currentYear = ref(new Date().getFullYear())
const currentMonth = ref(new Date().getMonth() + 1)

const noteForm = ref({
  content: '',
  images: '',
  tag_ids: [] as number[]
})

const uploadedImages = ref<string[]>([])
const isUploading = ref(false)

// 标签筛选和删除相关状态
const tagSortOrder = ref<'asc' | 'desc'>('desc') // 默认降序（数量多的在前）
const showDeleteTagConfirm = ref(false)
const deletingTag = ref<Tag | null>(null)
const deleteTagMode = ref<'tag-only' | 'tag-and-notes'>('tag-only')

// 快速输入相关状态
const quickNoteContent = ref('')
const quickNoteTagIds = ref<number[]>([])
const quickNoteTaskIds = ref<number[]>([])
const quickUploadedImages = ref<string[]>([])
const showTagDropdown = ref(false)
const showTaskDropdown = ref(false)
const isSending = ref(false)
const chatTextarea = ref<HTMLElement | null>(null)
const availableTasks = ref<any[]>([])
const dropdownStyle = ref({})
const mentionStartPos = ref(0)
const currentMentionType = ref<'tag' | 'task' | null>(null)
const mentionKeyword = ref('')
const selectedTags = ref<Tag[]>([])
const selectedTasks = ref<any[]>([])
const isApplyingHighlight = ref(false) // 防止递归标志

// 解析内容，将标签和任务高亮（用于显示）
const parsedContent = computed(() => {
  if (!quickNoteContent.value) return []
  
  const segments: Array<{ type: string; text: string; color?: string }> = []
  let lastIndex = 0
  const text = quickNoteContent.value
  
  // 匹配 #标签名 和 @任务名（以空格、换行或字符串结尾为界）
  const regex = /(#[^\s#@]+|@[^\s#@]+)/g
  let match
  
  while ((match = regex.exec(text)) !== null) {
    // 添加普通文本
    if (match.index > lastIndex) {
      segments.push({
        type: 'text',
        text: text.substring(lastIndex, match.index)
      })
    }
    
    // 添加高亮文本
    const matchedText = match[0]
    if (matchedText.startsWith('#')) {
      const tagName = matchedText.substring(1)
      const tag = tags.value.find(t => t.name === tagName)
      segments.push({
        type: 'highlight-tag',
        text: matchedText,
        color: tag?.color || '#4a90e2'
      })
    } else if (matchedText.startsWith('@')) {
      segments.push({
        type: 'highlight-task',
        text: matchedText
      })
    }
    
    lastIndex = match.index + matchedText.length
  }
  
  // 添加剩余文本
  if (lastIndex < text.length) {
    segments.push({
      type: 'text',
      text: text.substring(lastIndex)
    })
  }
  
  return segments
})

const filteredTags = computed(() => {
  if (!mentionKeyword.value) return tags.value
  return tags.value.filter(tag => 
    tag.name.toLowerCase().includes(mentionKeyword.value.toLowerCase())
  )
})

const filteredTasks = computed(() => {
  if (!mentionKeyword.value) return availableTasks.value
  return availableTasks.value.filter(task => 
    task.name.toLowerCase().includes(mentionKeyword.value.toLowerCase())
  )
})

const filteredNotes = computed(() => {
  if (selectedTagId.value === null) return notes.value
  return notes.value.filter(note => 
    note.tags.some(tag => tag.id === selectedTagId.value)
  )
})

const calendarDays = computed(() => {
  const firstDay = new Date(currentYear.value, currentMonth.value - 1, 1).getDay()
  const daysInMonth = new Date(currentYear.value, currentMonth.value, 0).getDate()
  
  const days: number[] = []
  for (let i = 0; i < firstDay; i++) {
    days.push(0)
  }
  for (let i = 1; i <= daysInMonth; i++) {
    days.push(i)
  }
  return days
})

const recordDays = computed(() => {
  return Object.keys(calendarData.value).length
})

// 排序后的标签列表
const sortedTags = computed(() => {
  const tagList = [...tags.value]
  return tagList.sort((a, b) => {
    const countA = getTagNoteCount(a.id)
    const countB = getTagNoteCount(b.id)
    return tagSortOrder.value === 'desc' ? countB - countA : countA - countB
  })
})

const hasRecord = (day: number) => {
  const dateStr = `${currentYear.value}-${String(currentMonth.value).padStart(2, '0')}-${String(day).padStart(2, '0')}`
  return !!calendarData.value[dateStr]
}

const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

const parseImages = (imagesStr: string | null | undefined): string[] => {
  if (!imagesStr) return []
  try {
    return JSON.parse(imagesStr)
  } catch {
    return []
  }
}

// 高亮笔记内容中的标签和任务
const highlightNoteContent = (content: string): string => {
  if (!content) return ''
  
  // 正则匹配 #标签名 和 @任务名（后面必须是空格、换行或字符串结尾）
  const tagRegex = /#([^\s#@]+)(?=\s|$)/g
  const taskRegex = /@([^\s#@]+)(?=\s|$)/g
  
  const replacements: Array<{ start: number; end: number; html: string }> = []
  
  // 收集所有需要高亮的位置
  let match
  while ((match = tagRegex.exec(content)) !== null) {
    const tagName = match[1]
    // 检查标签是否存在
    const tagExists = tags.value.some(t => t.name === tagName)
    const className = tagExists ? 'highlight-tag' : 'highlight-tag'
    replacements.push({
      start: match.index,
      end: match.index + match[0].length,
      html: `<span class="${className}">${escapeHtml(match[0])}</span>`
    })
  }
  
  while ((match = taskRegex.exec(content)) !== null) {
    replacements.push({
      start: match.index,
      end: match.index + match[0].length,
      html: `<span class="highlight-task">${escapeHtml(match[0])}</span>`
    })
  }
  
  // 按位置排序
  replacements.sort((a, b) => a.start - b.start)
  
  // 构建高亮后的 HTML
  let highlightedHTML = ''
  if (replacements.length > 0) {
    let lastIndex = 0
    
    for (const replacement of replacements) {
      // 添加普通文本
      highlightedHTML += escapeHtml(content.substring(lastIndex, replacement.start))
      // 添加高亮文本
      highlightedHTML += replacement.html
      lastIndex = replacement.end
    }
    // 添加剩余文本
    highlightedHTML += escapeHtml(content.substring(lastIndex))
  } else {
    highlightedHTML = escapeHtml(content)
  }
  
  return highlightedHTML
}

const loadNotes = async () => {
  try {
    const res: any = await noteAPI.getNotes(selectedTagId.value || undefined, showTrash.value)
    notes.value = res.notes || []
  } catch (err) {
    console.error('加载笔记失败', err)
  }
}

const loadTags = async () => {
  try {
    const res: any = await tagAPI.getTags()
    tags.value = res.tags || []
  } catch (err) {
    console.error('加载标签失败', err)
  }
}

const loadCalendar = async () => {
  try {
    const res: any = await noteAPI.getCalendar(currentYear.value, currentMonth.value)
    calendarData.value = res.calendar || {}
  } catch (err) {
    console.error('加载日历失败', err)
  }
}

const viewNoteDetail = (note: Note) => {
  viewingNote.value = note
  showDetailModal.value = true
}

const closeDetailModal = () => {
  showDetailModal.value = false
  viewingNote.value = null
}

const previewImage = (imageUrl: string) => {
  previewImageUrl.value = imageUrl
  showImagePreview.value = true
}

const closeImagePreview = () => {
  showImagePreview.value = false
  previewImageUrl.value = ''
}

const editNote = (note: Note) => {
  closeDetailModal() // 关闭详情弹窗
  editingNote.value = note
  noteForm.value = {
    content: note.content,
    images: note.images || '',
    tag_ids: note.tags.map(t => t.id)
  }
  // 解析已有图片
  uploadedImages.value = parseImages(note.images)
  showCreateModal.value = true
}

const openCreateModal = () => {
  editingNote.value = null
  noteForm.value = { content: '', images: '', tag_ids: [] }
  uploadedImages.value = []
  showCreateModal.value = true
}

const closeModal = () => {
  showCreateModal.value = false
  editingNote.value = null
  noteForm.value = { content: '', images: '', tag_ids: [] }
  uploadedImages.value = []
}

const handleImageUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const files = target.files
  if (!files || files.length === 0) return

  isUploading.value = true
  try {
    for (let i = 0; i < files.length; i++) {
      const file = files[i]
      const res: any = await noteAPI.uploadImage(file)
      // 后端返回的 res.url 已经是完整路径（如 /uploads/xxx.jpg）
      uploadedImages.value.push(res.url)
    }
  } catch (err) {
    console.error('上传图片失败', err)
    toast.error('上传图片失败，请重试')
  } finally {
    isUploading.value = false
    target.value = '' // 清空文件选择
  }
}

const removeImage = async (index: number) => {
  const url = uploadedImages.value[index]
  try {
    // url 已经是路径格式（如 /uploads/xxx.jpg）
    await noteAPI.deleteImage(url)
    uploadedImages.value.splice(index, 1)
  } catch (err) {
    console.error('删除图片失败', err)
    toast.error('删除图片失败')
  }
}

const saveNote = async () => {
  if (!noteForm.value.content.trim()) {
    toast.warning('请输入笔记内容')
    return
  }
  
  // 将图片数组转换为 JSON 字符串
  noteForm.value.images = uploadedImages.value.length > 0 
    ? JSON.stringify(uploadedImages.value) 
    : ''
  
  try {
    if (editingNote.value) {
      await noteAPI.updateNote(editingNote.value.id, noteForm.value)
    } else {
      await noteAPI.createNote(noteForm.value)
    }
    closeModal()
    loadNotes()
    loadCalendar()
  } catch (err) {
    console.error('保存笔记失败', err)
    toast.error('保存笔记失败，请重试')
  }
}

const deleteNote = async (id: number) => {
  if (!confirm(showTrash.value ? '确定永久删除这条笔记吗？此操作不可恢复！' : '确定删除这条笔记吗？')) return
  
  try {
    await noteAPI.deleteNote(id)
    closeDetailModal() // 关闭详情弹窗
    loadNotes()
    if (!showTrash.value) {
      loadCalendar()
    }
  } catch (err) {
    console.error('删除笔记失败', err)
  }
}

// 切换回收站视图
const toggleTrash = () => {
  showTrash.value = !showTrash.value
  selectedTagId.value = null // 切换时清除标签筛选
  loadNotes()
}

// 恢复笔记
const restoreNote = async (id: number) => {
  try {
    await noteAPI.restoreNote(id)
    loadNotes()
    loadCalendar()
    toast.success('笔记已恢复')
  } catch (err: any) {
    console.error('恢复笔记失败', err)
    toast.error(err.response?.data?.error || '恢复笔记失败')
  }
}

// 永久删除笔记
const permanentDeleteNote = async (id: number) => {
  if (!confirm('确定永久删除这条笔记吗？此操作不可恢复！')) return
  
  try {
    await noteAPI.deleteNote(id)
    loadNotes()
    toast.success('笔记已永久删除')
  } catch (err: any) {
    console.error('永久删除笔记失败', err)
    toast.error(err.response?.data?.error || '永久删除笔记失败')
  }
}

// 清空回收站
const confirmEmptyTrash = async () => {
  if (!confirm(`确定清空回收站吗？这将永久删除 ${notes.value.length} 条笔记，此操作不可恢复！`)) return
  
  try {
    const res: any = await noteAPI.emptyTrash()
    loadNotes()
    toast.success(res.message || '回收站已清空')
  } catch (err: any) {
    console.error('清空回收站失败', err)
    toast.error(err.response?.data?.error || '清空回收站失败')
  }
}

const changeMonth = (delta: number) => {
  currentMonth.value += delta
  if (currentMonth.value > 12) {
    currentMonth.value = 1
    currentYear.value++
  } else if (currentMonth.value < 1) {
    currentMonth.value = 12
    currentYear.value--
  }
  loadCalendar()
}

// 获取标签对应的笔记数量
const getTagNoteCount = (tagId: number) => {
  return notes.value.filter(note => 
    note.tags.some(tag => tag.id === tagId)
  ).length
}

// 选择标签进行筛选
const selectTagFilter = (tagId: number) => {
  if (selectedTagId.value === tagId) {
    // 再次点击同一标签，取消筛选
    selectedTagId.value = null
  } else {
    selectedTagId.value = tagId
  }
}

// 清除标签筛选
const clearTagFilter = () => {
  selectedTagId.value = null
}

// 切换排序顺序
const toggleSortOrder = () => {
  tagSortOrder.value = tagSortOrder.value === 'desc' ? 'asc' : 'desc'
}

// 显示删除标签确认弹窗
const showDeleteTagModal = (tag: Tag) => {
  deletingTag.value = tag
  deleteTagMode.value = 'tag-only' // 默认选择仅删除标签
  showDeleteTagConfirm.value = true
}

// 确认删除标签
const confirmDeleteTag = async () => {
  if (!deletingTag.value) return
  
  const deleteNotes = deleteTagMode.value === 'tag-and-notes'
  const tagName = deletingTag.value.name
  const noteCount = getTagNoteCount(deletingTag.value.id)
  
  // 二次确认
  const confirmMessage = deleteNotes
    ? `确定要删除标签 #${tagName} 及其关联的 ${noteCount} 条笔记吗？此操作不可恢复！`
    : `确定要删除标签 #${tagName} 吗？将从 ${noteCount} 条笔记中移除该标签引用。`
  
  if (!confirm(confirmMessage)) return
  
  try {
    await tagAPI.deleteTag(deletingTag.value.id, { delete_notes: deleteNotes })
    
    showDeleteTagConfirm.value = false
    deletingTag.value = null
    
    // 如果当前正在筛选被删除的标签，清除筛选
    if (selectedTagId.value === deletingTag.value?.id) {
      selectedTagId.value = null
    }
    
    // 重新加载数据
    await loadTags()
    await loadNotes()
    await loadCalendar()
    
    toast.success(deleteNotes ? '标签和笔记已删除' : '标签已删除，笔记内容已更新')
  } catch (err: any) {
    console.error('删除标签失败', err)
    toast.error(err.response?.data?.error || '删除标签失败，请重试')
  }
}

// 获取纯文本内容
const getTextContent = () => {
  return quickNoteContent.value.trim()
}

// 自动调整输入框高度
const autoResize = () => {
  if (!chatTextarea.value) return
  chatTextarea.value.style.height = 'auto'
  chatTextarea.value.style.height = chatTextarea.value.scrollHeight + 'px'
}

// 处理输入框点击
const handleTextareaClick = () => {
  // 点击时检查光标位置，如果在 # 或 @ 后面，显示下拉框
  setTimeout(() => {
    checkMentionTrigger()
  }, 0)
}

// 检查是否触发提及
const checkMentionTrigger = () => {
  if (!chatTextarea.value) return
  
  const selection = window.getSelection()
  if (!selection || selection.rangeCount === 0) return
  
  // 计算当前光标位置
  const range = selection.getRangeAt(0)
  const preCaretRange = range.cloneRange()
  preCaretRange.selectNodeContents(chatTextarea.value)
  preCaretRange.setEnd(range.startContainer, range.startOffset)
  const cursorPos = preCaretRange.toString().length
  
  const textBeforeCursor = quickNoteContent.value.substring(0, cursorPos)
  
  // 检查 # 标签
  const lastHashIndex = textBeforeCursor.lastIndexOf('#')
  if (lastHashIndex !== -1) {
    const textAfterHash = textBeforeCursor.substring(lastHashIndex + 1)
    if (!textAfterHash.includes(' ') && !textAfterHash.includes('\n') && !textAfterHash.includes('@')) {
      mentionStartPos.value = lastHashIndex
      currentMentionType.value = 'tag'
      mentionKeyword.value = textAfterHash
      showTagDropdown.value = true
      showTaskDropdown.value = false
      updateDropdownPosition()
      return
    }
  }
  
  // 检查 @ 任务
  const lastAtIndex = textBeforeCursor.lastIndexOf('@')
  if (lastAtIndex !== -1) {
    const textAfterAt = textBeforeCursor.substring(lastAtIndex + 1)
    if (!textAfterAt.includes(' ') && !textAfterAt.includes('\n') && !textAfterAt.includes('#')) {
      mentionStartPos.value = lastAtIndex
      currentMentionType.value = 'task'
      mentionKeyword.value = textAfterAt
      showTaskDropdown.value = true
      showTagDropdown.value = false
      updateDropdownPosition()
      return
    }
  }
  
  // 如果都不满足，关闭下拉框
  showTagDropdown.value = false
  showTaskDropdown.value = false
  currentMentionType.value = null
}

// 插入标签符号
const insertTag = () => {
  if (!chatTextarea.value) return
  
  const selection = window.getSelection()
  if (!selection || selection.rangeCount === 0) return
  
  // 计算当前光标位置
  const range = selection.getRangeAt(0)
  const preCaretRange = range.cloneRange()
  preCaretRange.selectNodeContents(chatTextarea.value)
  preCaretRange.setEnd(range.startContainer, range.startOffset)
  const cursorPos = preCaretRange.toString().length
  
  const textBefore = quickNoteContent.value.substring(0, cursorPos)
  const textAfter = quickNoteContent.value.substring(cursorPos)
  
  quickNoteContent.value = textBefore + '#' + textAfter
  
  // 设置光标位置和触发下拉框
  setTimeout(() => {
    if (chatTextarea.value) {
      const newPos = cursorPos + 1
      applyHighlight(chatTextarea.value, newPos)
      chatTextarea.value.focus()
      
      mentionStartPos.value = cursorPos
      currentMentionType.value = 'tag'
      mentionKeyword.value = ''
      showTagDropdown.value = true
      showTaskDropdown.value = false
      updateDropdownPosition()
    }
  }, 0)
}

// 插入任务符号
const insertTask = () => {
  if (!chatTextarea.value) return
  
  const selection = window.getSelection()
  if (!selection || selection.rangeCount === 0) return
  
  // 计算当前光标位置
  const range = selection.getRangeAt(0)
  const preCaretRange = range.cloneRange()
  preCaretRange.selectNodeContents(chatTextarea.value)
  preCaretRange.setEnd(range.startContainer, range.startOffset)
  const cursorPos = preCaretRange.toString().length
  
  const textBefore = quickNoteContent.value.substring(0, cursorPos)
  const textAfter = quickNoteContent.value.substring(cursorPos)
  
  quickNoteContent.value = textBefore + '@' + textAfter
  
  // 设置光标位置和触发下拉框
  setTimeout(() => {
    if (chatTextarea.value) {
      const newPos = cursorPos + 1
      applyHighlight(chatTextarea.value, newPos)
      chatTextarea.value.focus()
      
      mentionStartPos.value = cursorPos
      currentMentionType.value = 'task'
      mentionKeyword.value = ''
      showTaskDropdown.value = true
      showTagDropdown.value = false
      updateDropdownPosition()
    }
  }, 0)
}

// 更新下拉框位置
const updateDropdownPosition = () => {
  if (!chatTextarea.value) return
  
  // 简单定位：在输入框下方
  dropdownStyle.value = {
    top: '100%',
    left: '0',
    marginTop: '4px'
  }
}

// 处理输入
const handleInput = (e: Event) => {
  // 如果正在应用高亮，跳过
  if (isApplyingHighlight.value) {
    return
  }
  
  const target = e.target as HTMLElement
  
  // 获取纯文本内容
  const text = target.innerText || ''
  quickNoteContent.value = text
  
  // 保存光标位置
  const selection = window.getSelection()
  let cursorOffset = 0
  if (selection && selection.rangeCount > 0) {
    const range = selection.getRangeAt(0)
    const preCaretRange = range.cloneRange()
    preCaretRange.selectNodeContents(target)
    preCaretRange.setEnd(range.startContainer, range.startOffset)
    cursorOffset = preCaretRange.toString().length
  }
  
  // 应用高亮
  applyHighlight(target, cursorOffset)
  
  autoResize()
  checkMentionTrigger()
}

// 应用高亮显示
const applyHighlight = (element: HTMLElement, cursorOffset: number) => {
  const text = quickNoteContent.value
  
  // 如果文本为空，直接返回
  if (!text) {
    return
  }
  
  // 正则匹配 #标签名 和 @任务名（后面必须是空格、换行或字符串结尾）
  const tagRegex = /#([^\s#@]+)(?=\s|$)/g
  const taskRegex = /@([^\s#@]+)(?=\s|$)/g
  
  const replacements: Array<{ start: number; end: number; html: string }> = []
  
  // 收集所有需要高亮的位置
  let match
  while ((match = tagRegex.exec(text)) !== null) {
    const tagName = match[1]
    // 标签不存在时也高亮，但会在创建时自动创建
    replacements.push({
      start: match.index,
      end: match.index + match[0].length,
      html: `<span class="highlight-tag">${escapeHtml(match[0])}</span>`
    })
  }
  
  while ((match = taskRegex.exec(text)) !== null) {
    const taskName = match[1]
    // 检查任务是否存在
    const taskExists = availableTasks.value.some(t => t.name === taskName)
    const className = taskExists ? 'highlight-task' : 'highlight-task-invalid'
    replacements.push({
      start: match.index,
      end: match.index + match[0].length,
      html: `<span class="${className}">${escapeHtml(match[0])}</span>`
    })
  }
  
  // 按位置排序
  replacements.sort((a, b) => a.start - b.start)
  
  // 构建高亮后的 HTML
  let highlightedHTML = ''
  if (replacements.length > 0) {
    let lastIndex = 0
    
    for (const replacement of replacements) {
      // 添加普通文本
      highlightedHTML += escapeHtml(text.substring(lastIndex, replacement.start))
      // 添加高亮文本
      highlightedHTML += replacement.html
      lastIndex = replacement.end
    }
    // 添加剩余文本
    highlightedHTML += escapeHtml(text.substring(lastIndex))
  } else {
    highlightedHTML = escapeHtml(text)
  }
  
  // 只有当内容变化时才更新 DOM
  if (element.innerHTML !== highlightedHTML) {
    isApplyingHighlight.value = true
    element.innerHTML = highlightedHTML
    
    // 恢复光标位置
    restoreCursor(element, cursorOffset)
    
    // 重置标志
    setTimeout(() => {
      isApplyingHighlight.value = false
    }, 0)
  }
}

// HTML 转义
const escapeHtml = (text: string): string => {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

// 恢复光标位置
const restoreCursor = (element: HTMLElement, offset: number) => {
  const selection = window.getSelection()
  if (!selection) return
  
  let currentOffset = 0
  let found = false
  
  const findTextNode = (node: Node): boolean => {
    if (node.nodeType === Node.TEXT_NODE) {
      const textLength = node.textContent?.length || 0
      if (currentOffset + textLength >= offset) {
        const range = document.createRange()
        range.setStart(node, Math.min(offset - currentOffset, textLength))
        range.collapse(true)
        selection.removeAllRanges()
        selection.addRange(range)
        return true
      }
      currentOffset += textLength
    } else if (node.nodeType === Node.ELEMENT_NODE) {
      for (let i = 0; i < node.childNodes.length; i++) {
        if (findTextNode(node.childNodes[i]!)) {
          return true
        }
      }
    }
    return false
  }
  
  findTextNode(element)
}

// 处理键盘按下事件
const handleKeydown = (e: KeyboardEvent) => {
  // Enter 键发送消息（Shift+Enter 换行）
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendQuickNote()
  }
}

// 处理键盘释放事件
const handleKeyup = (e: KeyboardEvent) => {
  // ESC 键关闭下拉框
  if (e.key === 'Escape') {
    showTagDropdown.value = false
    showTaskDropdown.value = false
    currentMentionType.value = null
  }
}

// 处理粘贴事件
const handlePaste = (e: ClipboardEvent) => {
  e.preventDefault()
  
  // 获取纯文本
  const text = e.clipboardData?.getData('text/plain') || ''
  
  // 插入文本
  const selection = window.getSelection()
  if (selection && selection.rangeCount > 0) {
    const range = selection.getRangeAt(0)
    range.deleteContents()
    range.insertNode(document.createTextNode(text))
    range.collapse(false)
  }
  
  // 触发输入事件以更新高亮
  if (chatTextarea.value) {
    const event = new Event('input', { bubbles: true })
    chatTextarea.value.dispatchEvent(event)
  }
}

// 选择标签
const selectTag = (tag: Tag) => {
  if (!chatTextarea.value) return
  
  const text = quickNoteContent.value
  const selection = window.getSelection()
  if (!selection || selection.rangeCount === 0) return
  
  // 计算当前光标位置
  const range = selection.getRangeAt(0)
  const preCaretRange = range.cloneRange()
  preCaretRange.selectNodeContents(chatTextarea.value)
  preCaretRange.setEnd(range.startContainer, range.startOffset)
  const cursorPos = preCaretRange.toString().length
  
  const textBeforeCursor = text.substring(0, cursorPos)
  const textAfterCursor = text.substring(cursorPos)
  
  // 找到最近的 # 位置
  const lastHashIndex = textBeforeCursor.lastIndexOf('#')
  if (lastHashIndex === -1) return
  
  const textBeforeHash = text.substring(0, lastHashIndex)
  
  // 替换为 #标签名 （后面加空格）
  quickNoteContent.value = textBeforeHash + '#' + tag.name + ' ' + textAfterCursor
  
  // 添加到选中列表
  if (!quickNoteTagIds.value.includes(tag.id)) {
    quickNoteTagIds.value.push(tag.id)
    selectedTags.value.push(tag)
  }
  
  showTagDropdown.value = false
  currentMentionType.value = null
  
  // 更新显示并恢复光标
  setTimeout(() => {
    if (chatTextarea.value) {
      const newPos = textBeforeHash.length + tag.name.length + 2
      applyHighlight(chatTextarea.value, newPos)
      chatTextarea.value.focus()
      autoResize()
    }
  }, 0)
}

// 选择任务
const selectTask = (task: any) => {
  if (!chatTextarea.value) return
  
  const text = quickNoteContent.value
  const selection = window.getSelection()
  if (!selection || selection.rangeCount === 0) return
  
  // 计算当前光标位置
  const range = selection.getRangeAt(0)
  const preCaretRange = range.cloneRange()
  preCaretRange.selectNodeContents(chatTextarea.value)
  preCaretRange.setEnd(range.startContainer, range.startOffset)
  const cursorPos = preCaretRange.toString().length
  
  const textBeforeCursor = text.substring(0, cursorPos)
  const textAfterCursor = text.substring(cursorPos)
  
  // 找到最近的 @ 位置
  const lastAtIndex = textBeforeCursor.lastIndexOf('@')
  if (lastAtIndex === -1) return
  
  const textBeforeAt = text.substring(0, lastAtIndex)
  
  // 替换为 @任务名 （后面加空格）
  quickNoteContent.value = textBeforeAt + '@' + task.name + ' ' + textAfterCursor
  
  // 添加到选中列表
  if (!quickNoteTaskIds.value.includes(task.id)) {
    quickNoteTaskIds.value.push(task.id)
    selectedTasks.value.push(task)
  }
  
  showTaskDropdown.value = false
  currentMentionType.value = null
  
  // 更新显示并恢复光标
  setTimeout(() => {
    if (chatTextarea.value) {
      const newPos = textBeforeAt.length + task.name.length + 2
      applyHighlight(chatTextarea.value, newPos)
      chatTextarea.value.focus()
      autoResize()
    }
  }, 0)
}

// 触发图片上传
const triggerImageUpload = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  input.multiple = true
  input.onchange = async (e: Event) => {
    const target = e.target as HTMLInputElement
    const files = target.files
    if (!files || files.length === 0) return

    try {
      for (let i = 0; i < files.length; i++) {
        const file = files[i]
        const res: any = await noteAPI.uploadImage(file)
        // 后端返回的 res.url 已经是完整路径（如 /uploads/xxx.jpg）
        quickUploadedImages.value.push(res.url)
      }
    } catch (err) {
      console.error('上传图片失败', err)
      toast.error('上传图片失败，请重试')
    }
  }
  input.click()
}

// 移除快速上传的图片
const removeQuickImage = async (index: number) => {
  const url = quickUploadedImages.value[index]
  try {
    // url 已经是路径格式（如 /uploads/xxx.jpg）
    await noteAPI.deleteImage(url)
    quickUploadedImages.value.splice(index, 1)
  } catch (err) {
    console.error('删除图片失败', err)
  }
}

// 验证笔记内容中的任务引用
const validateTaskReferences = (content: string): { valid: boolean; invalidTasks: string[] } => {
  const taskRegex = /@([^\s#@]+)(?=\s|$)/g
  const invalidTasks: string[] = []
  let match
  
  while ((match = taskRegex.exec(content)) !== null) {
    const taskName = match[1]
    const taskExists = availableTasks.value.some(t => t.name === taskName)
    if (!taskExists) {
      invalidTasks.push(taskName)
    }
  }
  
  return {
    valid: invalidTasks.length === 0,
    invalidTasks
  }
}

// 从内容中提取标签名并自动创建不存在的标签
const extractAndCreateTags = async (content: string): Promise<number[]> => {
  const tagRegex = /#([^\s#@]+)(?=\s|$)/g
  const tagNames = new Set<string>()
  let match
  
  while ((match = tagRegex.exec(content)) !== null) {
    tagNames.add(match[1])
  }
  
  const tagIds: number[] = []
  
  for (const tagName of tagNames) {
    let tag = tags.value.find(t => t.name === tagName)
    
    if (!tag) {
      // 标签不存在，创建新标签
      try {
        const res: any = await tagAPI.createTag({ name: tagName, color: '#10b981' })
        tag = res.tag
        tags.value.push(tag)
      } catch (err) {
        console.error('创建标签失败', err)
        continue
      }
    }
    
    if (tag && !tagIds.includes(tag.id)) {
      tagIds.push(tag.id)
    }
  }
  
  return tagIds
}

// 发送快速笔记
const sendQuickNote = async () => {
  if (!quickNoteContent.value.trim() || isSending.value) return

  // 验证任务引用
  const validation = validateTaskReferences(quickNoteContent.value)
  if (!validation.valid) {
    toast.warning(`以下任务不存在：${validation.invalidTasks.map(t => '@' + t).join(', ')}。请先创建这些任务或删除引用。`, 5000)
    return
  }

  isSending.value = true
  try {
    // 提取并创建标签
    const extractedTagIds = await extractAndCreateTags(quickNoteContent.value)
    
    // 合并手动选择的标签ID和从内容提取的标签ID
    const allTagIds = [...new Set([...quickNoteTagIds.value, ...extractedTagIds])]
    
    const noteData = {
      content: quickNoteContent.value.trim(),
      images: quickUploadedImages.value.length > 0 
        ? JSON.stringify(quickUploadedImages.value) 
        : '',
      tag_ids: allTagIds,
      task_ids: quickNoteTaskIds.value
    }

    await noteAPI.createNote(noteData)
    
    // 清空输入
    quickNoteContent.value = ''
    quickNoteTagIds.value = []
    quickNoteTaskIds.value = []
    quickUploadedImages.value = []
    showTagDropdown.value = false
    showTaskDropdown.value = false
    currentMentionType.value = null
    
    // 重置输入框高度和内容
    if (chatTextarea.value) {
      chatTextarea.value.style.height = 'auto'
      chatTextarea.value.innerHTML = ''
    }
    
    // 重新加载笔记列表和标签列表
    await loadNotes()
    await loadCalendar()
    await loadTags()
  } catch (err) {
    console.error('发送笔记失败', err)
    toast.error('发送笔记失败，请重试')
  } finally {
    isSending.value = false
  }
}

// 加载可用任务列表
const loadAvailableTasks = async () => {
  try {
    const res: any = await taskAPI.getTasks({ archived: false })
    availableTasks.value = (res.tasks || []).filter((task: any) => task.status !== 3)
  } catch (err) {
    console.error('加载任务列表失败', err)
  }
}

watch(selectedTagId, () => {
  loadNotes()
})

onMounted(() => {
  loadNotes()
  loadTags()
  loadCalendar()
  loadAvailableTasks()
})
</script>

<style scoped>
/* 聊天框输入区域样式 */
.chat-input-container {
  margin-bottom: 20px;
  position: relative;
}

.chat-input-box {
  background: #f8f9fa;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  padding: 16px;
  transition: all 0.2s;
  position: relative;
}

.chat-input-box:focus-within {
  border-color: #4a90e2;
  background: #ffffff;
}

.chat-textarea {
  width: 100%;
  min-height: 60px;
  max-height: 200px;
  border: none;
  background: transparent;
  font-size: 14px;
  line-height: 1.6;
  resize: none;
  outline: none;
  font-family: inherit;
  color: #333;
  overflow-y: auto;
}

.chat-textarea::placeholder {
  color: #999;
}

.chat-textarea[contenteditable]:empty:before {
  content: attr(data-placeholder);
  color: #999;
  pointer-events: none;
}

/* 高亮样式 */
.chat-textarea :deep(.highlight-tag) {
  color: #10b981;
  font-weight: 600;
  background-color: #d1fae5;
  padding: 2px 6px;
  border-radius: 4px;
}

.chat-textarea :deep(.highlight-task) {
  color: #8b5cf6;
  font-weight: 600;
  background-color: #ede9fe;
  padding: 2px 6px;
  border-radius: 4px;
}

.chat-textarea :deep(.highlight-task-invalid) {
  color: #ef4444;
  font-weight: 600;
  background-color: #fee2e2;
  padding: 2px 6px;
  border-radius: 4px;
  text-decoration: line-through;
}

/* 笔记内容中的高亮样式 */
.note-content :deep(.highlight-tag),
.detail-content :deep(.highlight-tag) {
  color: #10b981;
  font-weight: 600;
  background-color: #d1fae5;
  padding: 2px 6px;
  border-radius: 4px;
}

.note-content :deep(.highlight-task),
.detail-content :deep(.highlight-task) {
  color: #8b5cf6;
  font-weight: 600;
  background-color: #ede9fe;
  padding: 2px 6px;
  border-radius: 4px;
}

.chat-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e5e7eb;
}

.toolbar-left {
  display: flex;
  gap: 8px;
}

.toolbar-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: #666;
  font-size: 16px;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.toolbar-btn:hover {
  background: #e5e7eb;
  color: #333;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.char-count {
  font-size: 12px;
  color: #999;
}

.send-btn {
  width: 40px;
  height: 40px;
  border: none;
  background: #d1d5db;
  color: white;
  font-size: 14px;
  cursor: not-allowed;
  border-radius: 8px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.send-btn.active {
  background: #10b981;
  cursor: pointer;
}

.send-btn.active:hover {
  background: #059669;
  transform: scale(1.05);
}

.send-btn:disabled {
  opacity: 0.6;
}

/* 快速选择器样式 */
.quick-selector {
  margin-top: 12px;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
}

/* 提及下拉框样式 */
.mention-dropdown {
  position: absolute;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 200px;
  max-width: 400px;
  max-height: 300px;
  overflow: hidden;
  z-index: 100;
}

.mention-list {
  max-height: 300px;
  overflow-y: auto;
  padding: 4px 0;
}

.mention-item {
  padding: 10px 16px;
  cursor: pointer;
  transition: background 0.2s;
  font-size: 14px;
}

.mention-item:hover {
  background: #f0f9ff;
}

.mention-empty {
  padding: 20px 16px;
  text-align: center;
  color: #999;
  font-size: 13px;
}

.selector-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: #f8f9fa;
  border-bottom: 1px solid #e5e7eb;
  font-size: 13px;
  font-weight: 500;
}

.chat-input-box {
  position: relative;
}

.selector-header button {
  background: none;
  border: none;
  font-size: 20px;
  color: #999;
  cursor: pointer;
  line-height: 1;
  padding: 0;
  width: 24px;
  height: 24px;
}

.selector-header button:hover {
  color: #333;
}

.selector-content {
  padding: 12px;
  max-height: 200px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.selector-item {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 6px 8px;
  border-radius: 6px;
  transition: background 0.2s;
}

.selector-item:hover {
  background: #f8f9fa;
}

.selector-item input[type="checkbox"] {
  cursor: pointer;
}

/* 快速上传图片预览 */
.quick-images-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e5e7eb;
}

.quick-image-item {
  position: relative;
  width: 80px;
  height: 80px;
  border-radius: 8px;
  overflow: hidden;
  border: 2px solid #e5e7eb;
}

.quick-image-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.remove-quick-image {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 20px;
  height: 20px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.remove-quick-image:hover {
  background: rgba(0, 0, 0, 0.8);
}

.calendar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.calendar-nav {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 18px;
  color: #666;
  padding: 4px;
}

.calendar-nav:hover {
  color: #333;
}

.calendar-month {
  font-size: 14px;
  font-weight: 500;
}

.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 4px;
}

.calendar-day-header {
  text-align: center;
  font-size: 12px;
  color: #999;
  padding: 8px 0;
}

.calendar-day {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.calendar-day.empty {
  background-color: transparent;
}

.calendar-day.no-record {
  background-color: #fff;
  border: 1px solid #e5e5e5;
}

.calendar-day.has-record {
  background-color: #b3d9ff;
  border: 1px solid #99ccff;
}

.calendar-day:not(.empty):hover {
  transform: scale(1.1);
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
  max-width: 600px;
}

.modal-content h2 {
  margin-bottom: 20px;
}

.modal-content textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  font-size: 14px;
  font-family: inherit;
  resize: vertical;
  margin-bottom: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 10px;
  font-weight: 500;
}

.tag-select {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.tag-checkbox {
  display: flex;
  align-items: center;
  gap: 5px;
  cursor: pointer;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.image-upload-area {
  margin-top: 10px;
}

.upload-btn {
  display: inline-block;
  padding: 8px 16px;
  background: #4a90e2;
  color: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.2s;
}

.upload-btn:hover {
  background: #357abd;
}

.uploaded-images {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 15px;
}

.image-preview {
  position: relative;
  width: 100px;
  height: 100px;
  border-radius: 6px;
  overflow: hidden;
  border: 2px solid #e5e5e5;
}

.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.remove-image {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 24px;
  height: 24px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  font-size: 18px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.remove-image:hover {
  background: rgba(0, 0, 0, 0.8);
}

.note-images {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 12px 0;
}

.note-image {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 6px;
  cursor: pointer;
  transition: transform 0.2s;
}

.note-image:hover {
  transform: scale(1.05);
}

.note-card {
  cursor: pointer;
  transition: all 0.2s;
}

.note-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

/* 笔记详情弹窗样式 */
.detail-modal {
  max-width: 700px;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid #e5e5e5;
}

.detail-header h2 {
  margin: 0;
}

.close-btn {
  background: none;
  border: none;
  font-size: 32px;
  color: #999;
  cursor: pointer;
  line-height: 1;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  color: #333;
}

.detail-body {
  margin-bottom: 20px;
}

.detail-date {
  color: #999;
  font-size: 13px;
  margin-bottom: 15px;
}

.detail-content {
  font-size: 15px;
  line-height: 1.8;
  color: #333;
  margin-bottom: 20px;
  white-space: pre-wrap;
  word-break: break-word;
}

.detail-images {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 20px;
}

.detail-image {
  width: 150px;
  height: 150px;
  object-fit: cover;
  border-radius: 8px;
  cursor: pointer;
  transition: transform 0.2s;
}

.detail-image:hover {
  transform: scale(1.05);
}

.detail-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 15px;
  border-top: 1px solid #e5e5e5;
}

.btn-danger {
  background-color: #e74c3c;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.2s;
}

.btn-danger:hover {
  background-color: #c0392b;
}

/* 图片预览弹窗样式 */
.image-preview-modal {
  background: rgba(0, 0, 0, 0.9);
}

.image-preview-content {
  position: relative;
  max-width: 90vw;
  max-height: 90vh;
  display: flex;
  align-items: center;
  justify-content: center;
}

.image-preview-content .close-btn {
  position: absolute;
  top: -50px;
  right: 0;
  color: white;
  font-size: 40px;
  z-index: 1001;
}

.image-preview-content img {
  max-width: 100%;
  max-height: 90vh;
  object-fit: contain;
  border-radius: 8px;
}

/* 标签筛选区域样式 */
.tag-filter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.sort-btn {
  background: none;
  border: 1px solid #e5e7eb;
  padding: 4px 12px;
  border-radius: 6px;
  font-size: 12px;
  color: #666;
  cursor: pointer;
  transition: all 0.2s;
}

.sort-btn:hover {
  background: #f3f4f6;
  color: #333;
  border-color: #d1d5db;
}

.tag-filter-list {
  max-height: 300px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.tag-filter-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: #f8f9fa;
  border: 2px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.tag-filter-item:hover {
  background: #e9ecef;
}

.tag-filter-item.active {
  background: #e3f2fd;
  border-color: #2196f3;
}

.tag-info {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
}

.tag-hash {
  font-size: 16px;
  font-weight: 600;
  color: #10b981;
}

.tag-name {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.tag-count {
  display: inline-block;
  background: #e5e7eb;
  color: #666;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
  margin-left: auto;
}

.tag-filter-item.active .tag-count {
  background: #90caf9;
  color: #1565c0;
}

.tag-delete-btn {
  width: 24px;
  height: 24px;
  border: none;
  background: #ef4444;
  color: white;
  border-radius: 50%;
  cursor: pointer;
  font-size: 18px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: all 0.2s;
  margin-left: 8px;
}

.tag-filter-item:hover .tag-delete-btn {
  opacity: 1;
}

.tag-delete-btn:hover {
  background: #dc2626;
  transform: scale(1.1);
}

.tag-filter-item.clear-filter {
  background: #fff3cd;
  border-color: #ffc107;
  justify-content: center;
}

.tag-filter-item.clear-filter:hover {
  background: #ffe69c;
}

/* 删除标签弹窗样式 */
.delete-tag-modal {
  max-width: 500px;
}

.delete-tag-info {
  background: #f8f9fa;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
  text-align: center;
}

.tag-name-display {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  margin-bottom: 12px;
}

.tag-name-display .tag-hash {
  color: #10b981;
}

.tag-note-count {
  font-size: 14px;
  color: #666;
}

.tag-note-count strong {
  color: #2196f3;
  font-size: 18px;
}

.delete-options {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
}

.delete-option {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.delete-option:hover {
  border-color: #cbd5e1;
  background: #f8f9fa;
}

.delete-option input[type="radio"] {
  margin-top: 4px;
  cursor: pointer;
}

.delete-option input[type="radio"]:checked ~ .option-content {
  color: #2196f3;
}

.option-content {
  flex: 1;
}

.option-title {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 6px;
  color: #333;
}

.option-desc {
  font-size: 13px;
  color: #666;
  line-height: 1.5;
}

.option-desc.warning {
  color: #dc2626;
  font-weight: 500;
}

/* 回收站相关样式 */
.header-actions {
  display: flex;
  gap: 10px;
}

.trash-note {
  opacity: 0.8;
  border-left: 4px solid #ef4444;
}

.trash-note:hover {
  opacity: 1;
}

.trash-info {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e5e7eb;
  font-size: 13px;
  color: #999;
}

.btn-restore {
  background-color: #10b981 !important;
  color: white !important;
}

.btn-restore:hover {
  background-color: #059669 !important;
}

.btn-permanent-delete {
  background-color: #ef4444 !important;
  color: white !important;
}

.btn-permanent-delete:hover {
  background-color: #dc2626 !important;
}
</style>
