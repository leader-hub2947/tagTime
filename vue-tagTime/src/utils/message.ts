import { createApp, h } from 'vue'
import Toast from '../components/Toast.vue'
import Modal from '../components/Modal.vue'

// Toast 实例
let toastInstance: any = null

const getToastInstance = () => {
  if (!toastInstance) {
    const container = document.createElement('div')
    document.body.appendChild(container)
    
    const app = createApp(Toast)
    toastInstance = app.mount(container)
  }
  return toastInstance
}

// Toast 方法
export const toast = {
  success: (message: string, duration?: number) => {
    getToastInstance().success(message, duration)
  },
  error: (message: string, duration?: number) => {
    getToastInstance().error(message, duration)
  },
  warning: (message: string, duration?: number) => {
    getToastInstance().warning(message, duration)
  },
  info: (message: string, duration?: number) => {
    getToastInstance().info(message, duration)
  }
}

// 确认对话框
export const confirm = (
  message: string,
  title: string = '确认操作',
  options?: {
    type?: 'confirm' | 'warning' | 'danger'
    confirmText?: string
    cancelText?: string
  }
): Promise<boolean> => {
  return new Promise((resolve) => {
    const container = document.createElement('div')
    document.body.appendChild(container)
    
    let modalRef: any = null
    
    const app = createApp({
      render() {
        return h(Modal, {
          ref: (el: any) => {
            modalRef = el
          },
          title: title,
          message: message,
          type: options?.type || 'confirm',
          confirmText: options?.confirmText || '确定',
          cancelText: options?.cancelText || '取消',
          onConfirm: () => {
            resolve(true)
            cleanup()
          },
          onCancel: () => {
            resolve(false)
            cleanup()
          },
          onClose: () => {
            resolve(false)
            cleanup()
          }
        })
      }
    })
    
    app.mount(container)
    
    // 显示模态框
    setTimeout(() => {
      if (modalRef && modalRef.show) {
        modalRef.show()
      }
    }, 0)
    
    const cleanup = () => {
      setTimeout(() => {
        app.unmount()
        document.body.removeChild(container)
      }, 300)
    }
  })
}

// 替代原生 alert
export const alert = (message: string, type: 'success' | 'error' | 'warning' | 'info' = 'info') => {
  toast[type](message)
}

// 导出默认对象
export default {
  toast,
  confirm,
  alert
}
