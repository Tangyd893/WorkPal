import { Link } from 'react-router-dom'

export default function NotFoundPage() {
  return (
    <main className="not-found-shell">
      <section className="module-surface not-found-panel">
        <span className="eyebrow">404</span>
        <h1>页面不存在</h1>
        <p>当前地址没有匹配到 WorkPal 页面。</p>
        <Link className="primary-button button-link" to="/workspace/overview">
          返回工作台
        </Link>
      </section>
    </main>
  )
}
