import { useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { FilesList } from '../../components/Files/FilesList'
import { Button } from 'primereact/button'

export default function Files() {
  const [searchParams, setSearchParams] = useSearchParams()

  const currentPath = searchParams.get('path') || ''

  useEffect(() => {
    // Sync currentPath from URL
  }, [currentPath])

  const handleRefreshClick = () => {
    setSearchParams({})
  }

  const handleNavigateToRoot = () => {
    setSearchParams({})
  }

  const handleNavigateToPath = (path: string) => {
    if (path) {
      setSearchParams({ path })
    } else {
      setSearchParams({})
    }
  }

  return (
    <div className="p-4">
      <div className="flex justify-content-between align-items-center mb-4">
        <h1>Файлы и папки</h1>
        <div className="flex gap-2">
          {currentPath && (
            <Button
              label="Корень"
              icon="pi pi-home"
              onClick={handleNavigateToRoot}
              text
            />
          )}
          <Button
            label="Обновить"
            icon="pi pi-refresh"
            onClick={handleRefreshClick}
          />
        </div>
      </div>
      <FilesList
        currentPath={currentPath}
        onNavigate={handleNavigateToPath}
      />
    </div>
  )
}
