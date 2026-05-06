import { useEffect, useMemo, useState } from 'react'
import type { AppTranslations } from '../i18n'
import type { WorkspaceSection } from '../types/workspace'

interface ModuleSwitcherProps {
  open: boolean
  activeSection: WorkspaceSection
  sections: WorkspaceSection[]
  labels: AppTranslations
  onNavigate: (section: WorkspaceSection) => void
  onClose: () => void
}

export default function ModuleSwitcher({ open, activeSection, sections, labels, onNavigate, onClose }: ModuleSwitcherProps) {
  const [query, setQuery] = useState('')

  useEffect(() => {
    if (!open) {
      setQuery('')
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

  const filteredSections = useMemo(() => {
    const normalizedQuery = query.trim().toLocaleLowerCase()
    if (!normalizedQuery) {
      return sections
    }

    return sections.filter((section) => labels.navigation[section].toLocaleLowerCase().includes(normalizedQuery))
  }, [labels.navigation, query, sections])

  if (!open) {
    return null
  }

  return (
    <div className="dialog-scrim">
      <div className="dialog-panel module-switcher" role="dialog" aria-modal="true" aria-labelledby="module-switcher-title">
        <div className="dialog-header">
          <div>
            <h3 id="module-switcher-title">{labels.shell.moduleSwitcher}</h3>
            <p>{labels.shell.moduleSwitcherHint}</p>
          </div>
        </div>
        <input
          type="search"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          autoFocus
          aria-label={labels.common.search}
          placeholder={labels.common.search}
        />
        <div className="choice-list">
          {filteredSections.map((section) => (
            <button
              key={section}
              type="button"
              className={section === activeSection ? 'choice-card selected' : 'choice-card'}
              aria-pressed={section === activeSection}
              onClick={() => {
                onNavigate(section)
                onClose()
              }}
            >
              <div>
                <strong>{labels.navigation[section]}</strong>
                <span>{section}</span>
              </div>
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
