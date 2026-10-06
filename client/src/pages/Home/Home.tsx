import { MetricsBar } from '../../components/MetricsBar/MetricsBar'

export default function Home() {
  console.log('render');
  return (
    <div>
      <MetricsBar />
      <h1>Home</h1>
      <p>Home component</p>
    </div>
  )
}
