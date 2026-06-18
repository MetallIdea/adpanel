import { useNavigate } from 'react-router-dom'
import { SitesList } from '../../components/Sites/SitesList'
import { Button } from 'primereact/button'

export default function Sites() {
  const navigate = useNavigate()

  const handleCreateClick = () => {
    navigate('/sites/create')
  }

  return (
    <div className="p-4">
      <div className="flex justify-content-between align-items-center mb-4">
        <h1>Управление сайтами</h1>
        <Button
          label="Создать сайт"
          icon="pi pi-plus"
          onClick={handleCreateClick}
        />
      </div>
      <SitesList />
    </div>
  )
}
