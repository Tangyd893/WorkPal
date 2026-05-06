import { useEffect } from 'react'
import { useToastStore, type ToastMessage } from '../hooks/useToastStore'

interface ToastItemProps {
  toast: ToastMessage
  onClose: (id: string) => void
}

function ToastItem({ toast, onClose }: ToastItemProps) {
  useEffect(() => {
    const timer = window.setTimeout(() => onClose(toast.id), toast.durationMs ?? 4000)
    return () => window.clearTimeout(timer)
  }, [onClose, toast.durationMs, toast.id])

  return (
    <div className={`toast-item ${toast.type}`} role={toast.type === 'error' ? 'alert' : 'status'}>
      <span>{toast.message}</span>
      <button type="button" className="toast-close" aria-label="关闭通知" onClick={() => onClose(toast.id)}>
        ×
      </button>
    </div>
  )
}

export default function ToastViewport() {
  const toasts = useToastStore((state) => state.toasts)
  const removeToast = useToastStore((state) => state.removeToast)

  if (toasts.length === 0) {
    return null
  }

  return (
    <div className="toast-viewport" aria-live="polite" aria-relevant="additions removals">
      {toasts.map((toast) => (
        <ToastItem key={toast.id} toast={toast} onClose={removeToast} />
      ))}
    </div>
  )
}
