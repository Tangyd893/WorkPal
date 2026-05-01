import { afterEach, describe, expect, it, vi } from 'vitest'
import { translations } from '../../i18n'
import { changeFile, render } from '../../test/render'
import type { SharedDocument } from '../../types/workspace'
import FilesPanel from './FilesPanel'

let view: ReturnType<typeof render> | null = null

afterEach(() => {
  view?.unmount()
  view = null
})

const baseDocument: SharedDocument = {
  id: 'file-1',
  title: 'Roadmap.pdf',
  summary: '12 KB',
  category: 'Upload',
  ownerUsername: 'admin',
  updatedAt: '2026-05-01T08:00:00Z',
  status: 'ready',
  sharedCount: 1,
  source: 'custom',
  fileId: 1,
  attachmentName: 'Roadmap.pdf',
  attachmentUrl: '/api/v1/files/1/download',
  downloadPath: '/api/v1/files/1/download',
}

function renderFilesPanel(documents: SharedDocument[] = [], onUpload = vi.fn().mockResolvedValue(undefined)) {
  view = render(
    <FilesPanel
      documents={documents}
      text={translations.en}
      getDisplayName={(username) => username}
      uploading={false}
      onUpload={onUpload}
      onDelete={vi.fn()}
      onShare={vi.fn()}
    />,
  )
  return { onUpload }
}

describe('FilesPanel', () => {
  it('marks upload controls and empty states with useful semantics', () => {
    renderFilesPanel()

    const fileInput = view?.container.querySelector<HTMLInputElement>('input[type="file"]')
    const hint = view?.container.querySelector('.banner-info')
    const empty = view?.container.querySelector('.empty-panel')

    expect(fileInput?.getAttribute('aria-label')).toBe(translations.en.files.uploadAction)
    expect(hint?.getAttribute('role')).toBe('status')
    expect(empty?.getAttribute('role')).toBe('status')
  })

  it('passes the selected file to the upload callback', async () => {
    const { onUpload } = renderFilesPanel()
    const fileInput = view?.container.querySelector<HTMLInputElement>('input[type="file"]')
    const file = new File(['hello'], 'hello.txt', { type: 'text/plain' })

    await changeFile(fileInput as HTMLInputElement, file)

    expect(onUpload).toHaveBeenCalledWith(file)
  })

  it('opens document links with noopener protection', () => {
    renderFilesPanel([baseDocument])

    const openLink = view?.container.querySelector<HTMLAnchorElement>('a.button-link')

    expect(openLink?.target).toBe('_blank')
    expect(openLink?.rel).toContain('noreferrer')
    expect(openLink?.rel).toContain('noopener')
  })
})
