import { useMemo, useState } from 'react'
import type { AppTranslations } from '../i18n'
import type { Locale, Notification, ThemeMode, WorkspaceSection } from '../types/workspace'

interface TopbarProps {
  activeSection: WorkspaceSection
  formattedDate: string
  locale: Locale
  theme: ThemeMode
  labels: AppTranslations
  notifications: Notification[]
  unreadCount: number
  onMarkRead?: (id: number) => void
  onMarkAllRead?: () => void
  onLocaleChange: (locale: Locale) => void
  onThemeChange: (theme: ThemeMode) => void
  onOpenSettings: () => void
  onLogout: () => void
}

export default function Topbar({
  activeSection,
  formattedDate,
  locale,
  theme,
  labels,
  notifications,
  unreadCount,
  onMarkRead,
  onMarkAllRead,
  onLocaleChange,
  onThemeChange,
  onOpenSettings,
  onLogout,
}: TopbarProps) {
  const [notificationsOpen, setNotificationsOpen] = useState(false)
  const notificationLabel = useMemo(
    () => `${labels.shell.notifications}${unreadCount > 0 ? ` (${unreadCount})` : ''}`,
    [labels.shell.notifications, unreadCount],
  )

  return (
    <header className="workspace-topbar">
      <div>
        <span className="eyebrow">
          {labels.shell.datePrefix} {formattedDate}
        </span>
        <h1>{labels.navigation[activeSection]}</h1>
        <p>{labels.shell.liveData}</p>
      </div>

      <div className="topbar-actions">
        <div className="segmented-control">
          <button
            type="button"
            className={locale === 'en' ? 'segment-button active' : 'segment-button'}
            aria-pressed={locale === 'en'}
            onClick={() => onLocaleChange('en')}
          >
            English
          </button>
          <button
            type="button"
            className={locale === 'zh-CN' ? 'segment-button active' : 'segment-button'}
            aria-pressed={locale === 'zh-CN'}
            onClick={() => onLocaleChange('zh-CN')}
          >
            简体中文
          </button>
        </div>

        <div className="segmented-control">
          <button
            type="button"
            className={theme === 'light' ? 'segment-button active' : 'segment-button'}
            aria-pressed={theme === 'light'}
            onClick={() => onThemeChange('light')}
          >
            {labels.settings.light}
          </button>
          <button
            type="button"
            className={theme === 'dark' ? 'segment-button active' : 'segment-button'}
            aria-pressed={theme === 'dark'}
            onClick={() => onThemeChange('dark')}
          >
            {labels.settings.dark}
          </button>
        </div>

        <div className="notification-shell">
          <button
            type="button"
            className="icon-button"
            aria-label={notificationLabel}
            aria-expanded={notificationsOpen}
            onClick={() => setNotificationsOpen((current) => !current)}
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M6 10a6 6 0 0 1 12 0v5l2 3H4l2-3v-5zm4 10h4" />
            </svg>
            {unreadCount > 0 ? <span className="notification-badge">{unreadCount}</span> : null}
          </button>
          {notificationsOpen ? (
            <div className="notification-popover" role="status">
              <div className="notification-popover-header">
                <strong>{labels.shell.notifications}</strong>
                {unreadCount > 0 && onMarkAllRead ? (
                  <button
                    type="button"
                    className="text-button"
                    onClick={() => { void onMarkAllRead() }}
                  >
                    {labels.shell.markAllRead}
                  </button>
                ) : null}
              </div>
              {notifications.length > 0 ? (
                notifications.map((item) => (
                  <div
                    key={item.id}
                    className={item.is_read ? 'notification-item is-read' : 'notification-item is-unread'}
                    onClick={() => {
                      if (!item.is_read && onMarkRead) {
                        void onMarkRead(item.id)
                      }
                    }}
                  >
                    <span className="notification-title">{item.title}</span>
                    <span className="notification-content">{item.content}</span>
                    <span className="notification-time">
                      {new Date(item.created_at).toLocaleString()}
                    </span>
                  </div>
                ))
              ) : (
                <span>{labels.shell.noNotifications}</span>
              )}
            </div>
          ) : null}
        </div>

        <button type="button" className="secondary-button" onClick={onOpenSettings}>
          {labels.shell.preferences}
        </button>
        <button type="button" className="secondary-button" onClick={onLogout}>
          {labels.shell.signOut}
        </button>
      </div>
    </header>
  )
}
