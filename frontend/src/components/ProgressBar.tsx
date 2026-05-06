interface ProgressBarProps {
  value: number
  label: string
}

export default function ProgressBar({ value, label }: ProgressBarProps) {
  const normalizedValue = Math.max(0, Math.min(100, Math.round(value)))

  return (
    <div className="progress-shell" role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={normalizedValue}>
      <div className="progress-label">
        <span>{label}</span>
        <strong>{normalizedValue}%</strong>
      </div>
      <div className="progress-track">
        <span style={{ width: `${normalizedValue}%` }} />
      </div>
    </div>
  )
}
