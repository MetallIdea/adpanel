import { useNavigate } from 'react-router-dom'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Card } from 'primereact/card'
import { Button } from 'primereact/button'
import { useGetSitesQuery } from '../../services/sitesApi'

type Site = {
  id: number
  name: string
  url: string
  port: number
  visits: number
}

export function SitesList() {
  const navigate = useNavigate()
  const { data, isLoading } = useGetSitesQuery()

  const handleEdit = (id: number) => {
    navigate(`/sites/${id}/edit`)
  }

  return (
    <div className="p-4">
      <Card title="Сайты" className="max-w-4xl mx-auto">
        <DataTable value={data} loading={isLoading} scrollable scrollHeight="400px">
          <Column field="id" header="ID" style={{ width: '80px' }}></Column>
          <Column field="name" header="Название"></Column>
          <Column field="url" header="URL"></Column>
          <Column field="port" header="Порт" style={{ width: '100px' }}></Column>
          <Column field="visits" header="Посещения" style={{ width: '120px' }}></Column>
          <Column
            header="Действия"
            body={(site: Site) => (
              <Button
                label="Редактировать"
                icon="pi pi-pencil"
                onClick={() => handleEdit(site.id)}
                className="p-button-sm p-button-text"
              />
            )}
            style={{ width: '140px' }}
          ></Column>
        </DataTable>
      </Card>
    </div>
  )
}
