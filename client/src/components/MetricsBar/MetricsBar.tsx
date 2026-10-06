import { useEffect, useState, useCallback } from 'react'
import { MeterGroup } from 'primereact/metergroup'
import { Button } from 'primereact/button'
import { getServerStatus, type ServerStatus } from '../../services/serverApi'
import styles from './MetricsBar.module.css'

const METER_CONFIGS = [
  { key: 'cpu', label: 'CPU', color: '#4caf50' },
  { key: 'memory', label: 'RAM', color: '#2196f3' },
  { key: 'disk', label: 'Disk', color: '#ff9800' },
] as const

export function MetricsBar() {
  const [status, setStatus] = useState<ServerStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchStatus = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await getServerStatus()
      setStatus(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch status')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchStatus();
  }, [])

  useEffect(() => {
    const interval = setInterval(fetchStatus, 60000)
    return () => clearInterval(interval)
  }, [fetchStatus])

  const values = METER_CONFIGS.map((config) => ({
    label: config.label,
    value: status?.[config.key]?.usage_percent ?? 0,
    color: config.color,
  }))

  const formatValue = (key: 'cpu' | 'memory' | 'disk'): string => {
    const data = status?.[key]
    if (!data) return '—'
    if (key === 'cpu') return `${data.usage_percent.toFixed(1)}%`
    const memData = data as { used_mb: number; total_mb: number }
    return `${memData.used_mb} / ${memData.total_mb} MB`
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h2>Метрики сервера</h2>
        <Button
          icon="pi pi-refresh"
          rounded
          text
          onClick={fetchStatus}
          disabled={loading}
          title="Обновить"
        />
      </div>

      {error && (
        <div className={styles.error}>
          {error}
        </div>
      )}

      <MeterGroup
        values={values}
        min={0}
        max={100}
      />

      <div className={styles.details}>
        {METER_CONFIGS.map((config) => (
          <div key={config.key} className={styles.detailItem}>
            <span className={styles.detailLabel}>{config.label}:</span>
            <span className={styles.detailValue}>{formatValue(config.key)}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
