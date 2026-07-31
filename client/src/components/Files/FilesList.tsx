import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Card } from 'primereact/card'
import { Badge } from 'primereact/badge'
import { useGetFilesQuery } from '../../services/filesApi'

type File = {
  id: number
  name: string
  type: 'file' | 'folder'
  size: number
  path: string
  createdAt: string
}

export function FilesList() {
  const { data, isLoading } = useGetFilesQuery()

  const typeBodyTemplate = (file: File) => {
    const typeColors = {
      file: 'p-secondary',
      folder: 'p-info',
    }

    const icon = file.type === 'folder' ? (
      <span className="pi pi-folder text-xl mr-2"></span>
    ) : (
      <span className="pi pi-file text-xl mr-2"></span>
    )

    return (
      <div className="flex align-items-center">
        {icon}
        <Badge value={file.type} className={typeColors[file.type]}></Badge>
      </div>
    )
  }

  const sizeBodyTemplate = (file: File) => {
    if (file.type === 'folder') return '-'
    if (file.size < 1024) return `${file.size} B`
    if (file.size < 1024 * 1024) return `${(file.size / 1024).toFixed(2)} KB`
    return `${(file.size / (1024 * 1024)).toFixed(2)} MB`
  }

  return (
    <div className="p-4">
      <Card title="Файлы и папки" className="max-w-6xl mx-auto">
        <DataTable value={data} loading={isLoading} responsiveLayout="scroll">
          <Column field="name" header="Название" style={{ width: '30%' }}></Column>
          <Column
            field="type"
            header="Тип"
            body={typeBodyTemplate}
            style={{ width: '15%' }}
          ></Column>
          <Column
            field="size"
            header="Размер"
            body={sizeBodyTemplate}
            style={{ width: '15%' }}
          ></Column>
          <Column field="path" header="Путь" style={{ width: '30%' }}></Column>
          <Column
            field="createdAt"
            header="Создано"
            style={{ width: '10%' }}
          ></Column>
        </DataTable>
      </Card>
    </div>
  )
}
