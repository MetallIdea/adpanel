import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Card } from 'primereact/card'
import { Badge } from 'primereact/badge'
import { get } from '../../utils/requests'

interface Site {
  id: number
  name: string
  url: string
  status: 'active' | 'inactive' | 'pending'
  visits: number
}

const sitesData: Site[] = [
  { id: 1, name: 'Main Site', url: 'https://example.com', status: 'active', visits: 12500 },
  { id: 2, name: 'Blog', url: 'https://blog.example.com', status: 'active', visits: 8300 },
  { id: 3, name: 'Admin Panel', url: 'https://admin.example.com', status: 'inactive', visits: 2100 },
  { id: 4, name: 'Documentation', url: 'https://docs.example.com', status: 'pending', visits: 5600 },
  { id: 5, name: 'API Gateway', url: 'https://api.example.com', status: 'active', visits: 45000 },
]

export function SitesList() {
  const statusBodyTemplate = (site: Site) => {
    const statusColors = {
      active: 'p-success',
      inactive: 'p-danger',
      pending: 'p-warning',
    }

    return (
      <Badge value={site.status} className={statusColors[site.status]}></Badge>
    )
  }

  return (
    <div className="p-4">
      <Card title="Сайты" className="max-w-4xl mx-auto">
        <DataTable value={sitesData} loading={true}>
          <Column field="id" header="ID" style={{ width: '80px' }}></Column>
          <Column field="name" header="Название"></Column>
          <Column field="url" header="URL"></Column>
          <Column
            field="status"
            header="Статус"
            body={statusBodyTemplate}
            style={{ width: '120px' }}
          ></Column>
          <Column field="visits" header="Посещения" style={{ width: '120px' }}></Column>
        </DataTable>
      </Card>
    </div>
  )
}
