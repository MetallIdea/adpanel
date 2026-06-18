import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Card } from 'primereact/card'
import { Badge } from 'primereact/badge'
import { useGetSitesQuery } from '../../services/sitesApi'

type Site = {
  id: number
  name: string
  url: string
  status: 'active' | 'inactive' | 'pending'
  visits: number
}

export function SitesList() {
  const { data, isLoading } = useGetSitesQuery()

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
        <DataTable value={data} loading={isLoading}>
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
