import { useCallback, useEffect, useRef, useState } from 'react'
import { workpalApi } from '../api/workpal'
import type { Notification } from '../types/workspace'

export interface UseNotificationsResult {
  notifications: Notification[]
  unreadCount: number
  loading: boolean
  markRead: (id: number) => Promise<void>
  markAllRead: () => Promise<void>
}

export function useNotifications(pollIntervalMs = 30000): UseNotificationsResult {
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [loading, setLoading] = useState(true)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const fetchNotifications = useCallback(async () => {
    try {
      const [items, unread] = await Promise.all([
        workpalApi.listNotifications(),
        workpalApi.getUnreadNotificationCount(),
      ])
      setNotifications(items)
      setUnreadCount(unread.count)
    } catch {
      // 通知服务不可用时静默失败
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchNotifications()
    pollRef.current = setInterval(fetchNotifications, pollIntervalMs)
    return () => {
      if (pollRef.current) {
        clearInterval(pollRef.current)
      }
    }
  }, [fetchNotifications, pollIntervalMs])

  const markRead = useCallback(async (id: number) => {
    await workpalApi.markNotificationRead(id)
    setNotifications((prev) =>
      prev.map((n) => (n.id === id ? { ...n, is_read: true } : n)),
    )
    setUnreadCount((prev) => Math.max(0, prev - 1))
  }, [])

  const markAllRead = useCallback(async () => {
    await workpalApi.markAllNotificationsRead()
    setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })))
    setUnreadCount(0)
  }, [])

  return { notifications, unreadCount, loading, markRead, markAllRead }
}
