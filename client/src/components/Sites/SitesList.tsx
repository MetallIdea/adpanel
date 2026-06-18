import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Card } from 'primereact/card'
import { Badge } from 'primereact/badge'
import { useSelector } from 'react-redux'
import type { RootState } from '../../store/store'

interface Site {
  id: number
  name: string
  url: string
  status: 'active' | 'inactive' | 'pending'
  visits: number
}

export function SitesList() {
  const sites = useSelector((state: RootState) => state.sites.sites)
  const loading = useSelector((state: RootState) => state.sites.loading)

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
        <DataTable value={sites} loading={loading}>
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
