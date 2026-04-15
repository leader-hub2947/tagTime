<template>
  <teleport to="body">
    <transition name="modal-fade">
      <div v-if="visible" class="modal-overlay" @click="handleOverlayClick">
        <transition name="modal-slide">
          <div v-if="visible" class="modal-dialog" :class="[`modal-${type}`, sizeClass]">
            <div class="modal-header">
              <div class="modal-icon" v-if="type !== 'default'">
                <svg v-if="type === 'confirm'" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <path d="M12 9v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                <svg v-else-if="type === 'warning'" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                <svg v-else-if="type === 'danger'" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <path d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </div>
              <h3 class="modal-title">{{ title }}</h3>
              <button class="modal-close" @click="handleCancel">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <path d="M6 18L18 6M6 6l12 12" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </button>
            </div>
            
            <div class="modal-body">
              <p class="modal-message" v-if="message">{{ message }}</p>
              <slot></slot>
            </div>
            
            <div class="modal-footer" v-if="showFooter">
              <button class="btn btn-secondary" @click="handleCancel">
                {{ cancelText }}
              </button>
              <button 
                class="btn btn-primary" 
                :class="[`btn-${type}`]"
                @click="handleConfirm"
              >
                {{ confirmText }}
              </button>
            </div>
          </div>
        </transition>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  title?: string
  message?: string
  type?: 'default' | 'confirm' | 'warning' | 'danger'
  size?: 'small' | 'medium' | 'large'
  confirmText?: string
  cancelText?: string
  showFooter?: boolean
  closeOnOverlay?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  title: '提示',
  type: 'default',
  size: 'medium',
  confirmText: '确定',
  cancelText: '取消',
  showFooter: true,
  closeOnOverlay: true
})

const emit = defineEmits<{
  confirm: []
  cancel: []
  close: []
}>()

const visible = ref(false)

const sizeClass = computed(() => `modal-${props.size}`)

const show = () => {
  visible.value = true
}

const hide = () => {
  visible.value = false
  emit('close')
}

const handleConfirm = () => {
  emit('confirm')
  hide()
}

const handleCancel = () => {
  emit('cancel')
  hide()
}

const handleOverlayClick = () => {
  if (props.closeOnOverlay) {
    handleCancel()
  }
}

defineExpose({
  show,
  hide
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 20px;
  backdrop-filter: blur(4px);
}

.modal-dialog {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-small {
  width: 100%;
  max-width: 400px;
}

.modal-medium {
  width: 100%;
  max-width: 500px;
}

.modal-large {
  width: 100%;
  max-width: 700px;
}

.modal-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 24px 24px 16px;
  border-bottom: 1px solid #e5e5e5;
}

.modal-icon {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
}

.modal-confirm .modal-icon {
  color: #3b82f6;
}

.modal-warning .modal-icon {
  color: #f59e0b;
}

.modal-danger .modal-icon {
  color: #ef4444;
}

.modal-title {
  flex: 1;
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.modal-close {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  padding: 0;
  border: none;
  background: none;
  color: #999;
  cursor: pointer;
  transition: all 0.2s ease;
}

.modal-close:hover {
  color: #333;
  transform: scale(1.1);
}

.modal-close svg {
  width: 100%;
  height: 100%;
}

.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.modal-message {
  font-size: 15px;
  color: #666;
  line-height: 1.6;
  margin: 0;
  white-space: pre-wrap;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e5e5e5;
  background: #fafafa;
}

.btn {
  padding: 10px 24px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn:active {
  transform: scale(0.98);
}

.btn-secondary {
  background-color: #f0f0f0;
  color: #666;
}

.btn-secondary:hover {
  background-color: #e0e0e0;
  color: #333;
}

.btn-primary {
  background-color: #333;
  color: #fff;
}

.btn-primary:hover {
  background-color: #555;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.btn-confirm {
  background-color: #3b82f6;
}

.btn-confirm:hover {
  background-color: #2563eb;
}

.btn-warning {
  background-color: #f59e0b;
}

.btn-warning:hover {
  background-color: #d97706;
}

.btn-danger {
  background-color: #ef4444;
}

.btn-danger:hover {
  background-color: #dc2626;
}

/* 动画 */
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.3s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-slide-enter-active {
  animation: modalSlideIn 0.3s ease;
}

.modal-slide-leave-active {
  animation: modalSlideOut 0.3s ease;
}

@keyframes modalSlideIn {
  from {
    opacity: 0;
    transform: scale(0.9) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

@keyframes modalSlideOut {
  from {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
  to {
    opacity: 0;
    transform: scale(0.9) translateY(-20px);
  }
}

@media (max-width: 640px) {
  .modal-dialog {
    max-width: 100%;
    margin: 0;
  }
  
  .modal-header {
    padding: 20px 16px 12px;
  }
  
  .modal-body {
    padding: 20px 16px;
  }
  
  .modal-footer {
    padding: 12px 16px;
  }
}
</style>
