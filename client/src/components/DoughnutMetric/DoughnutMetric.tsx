import { useMemo } from 'react'
import { Chart } from 'primereact/chart'
import type { ChartData, ChartOptions } from 'chart.js'
import styles from './DoughnutMetric.module.css'

interface DoughnutMetricProps {
  percentage: number
  color: string
  label: string
}

export function DoughnutMetric({ percentage, color, label }: DoughnutMetricProps) {
  const data: ChartData<'doughnut'> = useMemo(
    () => ({
      labels: [label, ''],
      datasets: [
        {
          data: [percentage, 100 - percentage],
          backgroundColor: [color, 'var(--surface-border)'],
          borderWidth: 0,
        },
      ],
    }),
    [label, percentage, color]
  )

  const options: ChartOptions<'doughnut'> = useMemo(
    () => ({
      cutout: '75%',
      responsive: true,
      maintainAspectRatio: true,
      plugins: {
        legend: { display: false },
        tooltip: { enabled: false },
      },
      animation: { animate: true },
    }),
    []
  )

  return (
    <div className={styles.doughnutCard}>
      <div className={styles.doughnutWrapper}>
        <Chart type="doughnut" data={data} options={options} />
        <div className={styles.doughnutCenter}>
          <span className={styles.doughnutValue}>{percentage.toFixed(1)}%</span>
        </div>
      </div>
      <span className={styles.doughnutLabel}>{label}</span>
    </div>
  )
}
