import { useEffect } from 'react'
import type { AppTranslations } from '../../i18n'
import type { Locale, ThemeMode } from '../../types/workspace'

interface SettingsDrawerProps {
  open: boolean
  locale: Locale
  theme: ThemeMode
  soundEnabled: boolean
  compactMode: boolean
  text: AppTranslations
  onClose: () => void
  onLocaleChange: (locale: Locale) => void
  onThemeChange: (theme: ThemeMode) => void
  onSoundChange: (enabled: boolean) => void
  onCompactModeChange: (enabled: boolean) => void
  onReset: () => void
}

export default function SettingsDrawer({
  open,
  locale,
  theme,
  soundEnabled,
  compactMode,
  text,
  onClose,
  onLocaleChange,
  onThemeChange,
  onSoundChange,
  onCompactModeChange,
  onReset,
}: SettingsDrawerProps) {
  useEffect(() => {
    if (!open) {
      return undefined
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onClose, open])

  if (!open) {
    return null
  }

  return (
    <div className="drawer-scrim">
      <aside className="drawer-panel" role="dialog" aria-modal="true" aria-labelledby="settings-drawer-title">
        <div className="drawer-header">
          <div>
            <h3 id="settings-drawer-title">{text.settings.title}</h3>
            <p>{text.settings.subtitle}</p>
          </div>
          <button type="button" className="secondary-button" onClick={onClose}>
            {text.settings.close}
          </button>
        </div>

        <div className="settings-group">
          <div className="settings-copy">
            <strong>{text.settings.language}</strong>
            <span>{text.settings.languageHint}</span>
          </div>
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
        </div>

        <div className="settings-group">
          <div className="settings-copy">
            <strong>{text.settings.theme}</strong>
            <span>{text.settings.themeHint}</span>
          </div>
          <div className="segmented-control">
            <button
              type="button"
              className={theme === 'light' ? 'segment-button active' : 'segment-button'}
              aria-pressed={theme === 'light'}
              onClick={() => onThemeChange('light')}
            >
              {text.settings.light}
            </button>
            <button
              type="button"
              className={theme === 'dark' ? 'segment-button active' : 'segment-button'}
              aria-pressed={theme === 'dark'}
              onClick={() => onThemeChange('dark')}
            >
              {text.settings.dark}
            </button>
          </div>
        </div>

        <div className="settings-group">
          <div className="settings-copy">
            <strong>{text.settings.sound}</strong>
            <span>{text.settings.soundHint}</span>
          </div>
          <label className="toggle-row">
            <input type="checkbox" checked={soundEnabled} onChange={(event) => onSoundChange(event.target.checked)} />
            <span>{soundEnabled ? text.common.enabled : text.common.disabled}</span>
          </label>
        </div>

        <div className="settings-group">
          <div className="settings-copy">
            <strong>{text.settings.density}</strong>
            <span>{text.settings.densityHint}</span>
          </div>
          <div className="segmented-control">
            <button
              type="button"
              className={!compactMode ? 'segment-button active' : 'segment-button'}
              aria-pressed={!compactMode}
              onClick={() => onCompactModeChange(false)}
            >
              {text.settings.comfortable}
            </button>
            <button
              type="button"
              className={compactMode ? 'segment-button active' : 'segment-button'}
              aria-pressed={compactMode}
              onClick={() => onCompactModeChange(true)}
            >
              {text.settings.compact}
            </button>
          </div>
        </div>

        <div className="drawer-footer">
          <button type="button" className="secondary-button" onClick={onReset}>
            {text.settings.reset}
          </button>
        </div>
      </aside>
    </div>
  )
}
