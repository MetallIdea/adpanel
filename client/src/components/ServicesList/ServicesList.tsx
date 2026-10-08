import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Card } from 'primereact/card'
import { Badge } from 'primereact/badge'
import { useGetServicesQuery } from '../../services/servicesApi'

type Service = {
  id: number
  title: string
  description: string
  status: 'active' | 'inactive' | 'pending'
}

export function ServicesList() {
  const { data, isLoading } = useGetServicesQuery()

  const statusBodyTemplate = (service: Service) => {
    const statusColors = {
      active: 'p-success',
      inactive: 'p-danger',
      pending: 'p-warning',
    }

    return (
      <Badge value={service.status} className={statusColors[service.status]}></Badge>
    )
  }

  return (
    <div className="p-4">
      <Card title="Сервисы" className="max-w-4xl mx-auto">
        <DataTable value={data} loading={isLoading}>
          <Column field="id" header="ID" style={{ width: '80px' }}></Column>
          <Column field="title" header="Название"></Column>
          <Column field="description" header="Описание"></Column>
          <Column
            field="status"
            header="Статус"
            body={statusBodyTemplate}
            style={{ width: '120px' }}
          ></Column>
        </DataTable>
      </Card>
    </div>
  )
}
