import { useEffect, useMemo, useState } from 'react'
import type { AppTranslations } from '../i18n'
import type { WorkspaceSection, WorkspaceUser } from '../types/workspace'

interface SidebarProps {
  activeSection: WorkspaceSection
  userInfo: WorkspaceUser | null
  username: string | null
  labels: AppTranslations
  onNavigate: (section: WorkspaceSection) => void
  onOpenSettings: () => void
}

type NavigationGroupID = 'overview' | 'collaboration' | 'work' | 'projects' | 'knowledge' | 'assets'

interface NavigationItem {
  section: WorkspaceSection
  icon: string
}

interface NavigationGroup {
  id: NavigationGroupID
  items: NavigationItem[]
}

const navigationGroups: NavigationGroup[] = [
  { id: 'overview', items: [{ section: 'overview', icon: 'M4 5h16M4 12h10M4 19h7' }] },
  {
    id: 'collaboration',
    items: [
      { section: 'chat', icon: 'M5 6h14v9H8l-3 3V6z' },
      { section: 'directory', icon: 'M8 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8zm8 2a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM3 21a5 5 0 0 1 10 0m2 0a4 4 0 0 1 6 0' },
    ],
  },
  {
    id: 'work',
    items: [
      { section: 'tasks', icon: 'M5 12l4 4L19 6M5 20h14' },
      { section: 'schedule', icon: 'M7 3v4m10-4v4M4 9h16M5 5h14v16H5z' },
    ],
  },
  {
    id: 'projects',
    items: [
      { section: 'projects' as WorkspaceSection, icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z' },
    ],
  },
  {
    id: 'knowledge',
    items: [
      { section: 'docs' as WorkspaceSection, icon: 'M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z' },
      { section: 'calendar' as WorkspaceSection, icon: 'M7 3v4m10-4v4M4 9h16M5 5h14v16H5z' },
      { section: 'approvals' as WorkspaceSection, icon: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z' },
    ],
  },
  { id: 'assets', items: [{ section: 'files', icon: 'M5 4h6l2 3h6v13H5z' }] },
]

function findGroupID(section: WorkspaceSection): NavigationGroupID {
  return navigationGroups.find((group) => group.items.some((item) => item.section === section))?.id ?? 'overview'
}

function NavigationIcon({ path }: { path: string }) {
  return (
    <svg className="nav-icon" viewBox="0 0 24 24" aria-hidden="true">
      <path d={path} />
    </svg>
  )
}

export default function Sidebar({ activeSection, userInfo, username, labels, onNavigate, onOpenSettings }: SidebarProps) {
  const activeGroupID = findGroupID(activeSection)
  const [expandedGroups, setExpandedGroups] = useState<NavigationGroupID[]>(() => [activeGroupID])
  const [mobileNavOpen, setMobileNavOpen] = useState(false)

  useEffect(() => {
    setExpandedGroups((current) => (current.includes(activeGroupID) ? current : [...current, activeGroupID]))
  }, [activeGroupID])

  const displayName = userInfo?.nickname || username || labels.common.unavailable
  const accountName = userInfo?.username || username || 'guest'

  const expandedLookup = useMemo(() => new Set(expandedGroups), [expandedGroups])

  const toggleGroup = (groupID: NavigationGroupID) => {
    setExpandedGroups((current) =>
      current.includes(groupID) ? current.filter((item) => item !== groupID) : [...current, groupID],
    )
  }

  return (
    <aside className={mobileNavOpen ? 'workspace-sidebar mobile-open' : 'workspace-sidebar'}>
      <div className="brand-block">
        <div>
          <strong>{labels.common.workpal}</strong>
          <span>{labels.shell.subtitle}</span>
        </div>
        <button
          type="button"
          className="sidebar-toggle"
          aria-expanded={mobileNavOpen}
          aria-label={labels.common.open}
          onClick={() => setMobileNavOpen((current) => !current)}
        >
          <span />
          <span />
          <span />
        </button>
      </div>

      <nav className="workspace-nav" aria-label={labels.common.workpal}>
        {navigationGroups.map((group) => {
          const expanded = expandedLookup.has(group.id)

          return (
            <section key={group.id} className="nav-group">
              <button
                type="button"
                className={group.id === activeGroupID ? 'nav-group-button active' : 'nav-group-button'}
                aria-expanded={expanded}
                onClick={() => toggleGroup(group.id)}
              >
                <span>{labels.shell.navGroups[group.id]}</span>
                <span className="nav-chevron" aria-hidden="true" />
              </button>
              <div className={expanded ? 'nav-group-items expanded' : 'nav-group-items'}>
                {group.items.map((item) => (
                  <button
                    key={item.section}
                    type="button"
                    className={item.section === activeSection ? 'nav-button active' : 'nav-button'}
                    aria-current={item.section === activeSection ? 'page' : undefined}
                    onClick={() => {
                      onNavigate(item.section)
                      setMobileNavOpen(false)
                    }}
                  >
                    <NavigationIcon path={item.icon} />
                    <span>{labels.navigation[item.section]}</span>
                  </button>
                ))}
              </div>
            </section>
          )
        })}
      </nav>

      <div className="sidebar-footer">
        <div className="profile-card">
          <strong>{displayName}</strong>
          <span>@{accountName}</span>
        </div>
        <button type="button" className="secondary-button block-button" onClick={onOpenSettings}>
          {labels.shell.preferences}
        </button>
      </div>
    </aside>
  )
}
