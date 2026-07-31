import { useNavigate } from 'react-router-dom'
import { FilesList } from '../../components/Files/FilesList'
import { Button } from 'primereact/button'

export default function Files() {
  const navigate = useNavigate()

  const handleRefreshClick = () => {
    navigate('/files')
  }

  return (
    <div className="p-4">
      <div className="flex justify-content-between align-items-center mb-4">
        <h1>Файлы и папки</h1>
        <Button
          label="Обновить"
          icon="pi pi-refresh"
          onClick={handleRefreshClick}
        />
      </div>
      <FilesList />
    </div>
  )
}
