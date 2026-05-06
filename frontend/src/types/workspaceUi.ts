export interface ConfirmRequest {
  title: string
  message: string
  confirmText: string
  cancelText: string
  variant?: 'normal' | 'danger'
  onConfirm: () => Promise<void> | void
}
