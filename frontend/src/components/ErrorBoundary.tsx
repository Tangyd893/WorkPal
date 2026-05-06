import { Component, type ErrorInfo, type ReactNode } from 'react'

interface ErrorBoundaryProps {
  children: ReactNode
  resetKey?: string
  title?: string
  message?: string
}

interface ErrorBoundaryState {
  error: Error | null
}

export default class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { error: null }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { error }
  }

  componentDidUpdate(previousProps: ErrorBoundaryProps) {
    if (this.state.error && previousProps.resetKey !== this.props.resetKey) {
      this.setState({ error: null })
    }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('WorkPal rendering error.', error, info)
  }

  handleRetry = () => {
    this.setState({ error: null })
  }

  render() {
    if (!this.state.error) {
      return this.props.children
    }

    return (
      <section className="module-surface error-boundary-panel" role="alert">
        <h2>{this.props.title ?? '模块加载失败'}</h2>
        <p>{this.props.message ?? '当前模块渲染时出现异常，可以重试或切换到其他模块。'}</p>
        <button type="button" className="primary-button" onClick={this.handleRetry}>
          重试
        </button>
      </section>
    )
  }
}
