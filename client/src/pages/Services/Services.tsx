import { useNavigate } from 'react-router-dom'
import { ServicesList } from '../../components/ServicesList/ServicesList'
import { Button } from 'primereact/button'

export default function Services() {
  const navigate = useNavigate()

  const handleCreateClick = () => {
    navigate('/services/create')
  }

  return (
    <div className="p-4">
      <div className="flex justify-content-between align-items-center mb-4">
        <h1>Управление сервисами</h1>
        <Button
          label="Создать сервис"
          icon="pi pi-plus"
          onClick={handleCreateClick}
        />
      </div>
      <ServicesList />
    </div>
  )
}
