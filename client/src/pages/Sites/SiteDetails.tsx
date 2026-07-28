import { useParams } from 'react-router-dom'
import { useGetSiteByIdQuery } from '../../services/sitesApi'
import { Card } from 'primereact/card'

export default function SiteDetails() {
    const { id } = useParams<{ id: string }>()
    const siteId = Number(id)

    const { data: site, isLoading, error } = useGetSiteByIdQuery(siteId)

    if (isLoading) return <div>Загрузка...</div>
    if (error || !site) return <div>Сайт не найден</div>

    return (
        <div style={{ padding: '20px', maxWidth: '800px', margin: '0 auto' }}>
            <Card title={site.name} subTitle={`ID: ${site.id}`}>
                <p>
                    <strong>URL:</strong>{' '}
                    <a href={site.url} target="_blank" rel="noopener noreferrer">
                        {site.url}
                    </a>
                </p>
            </Card>
        </div>
    )
}
