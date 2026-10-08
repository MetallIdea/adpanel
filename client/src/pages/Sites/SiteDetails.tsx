import { useParams } from 'react-router-dom'
import { useGetSiteByIdQuery } from '../../services/sitesApi'
import { useGetNginxConfigQuery } from '../../services/sitesApi'
import { Card } from 'primereact/card'
import { Button } from 'primereact/button'
import { useState } from 'react'

export default function SiteDetails() {
    const { id } = useParams<{ id: string }>()
    const siteId = Number(id)

    const { data: site, isLoading, error } = useGetSiteByIdQuery(siteId)
    const { data: nginxData } = useGetNginxConfigQuery(siteId)
    const [showNginx, setShowNginx] = useState(false)

    if (isLoading) return <div>Загрузка...</div>
    if (error || !site) return <div>Сайт не найден</div>

    return (
        <div style={{ padding: '20px', maxWidth: '800px', margin: '0 auto' }}>
            <div className="flex justify-content-between align-items-center mb-4">
                <h2>Детали сайта</h2>
                <Button
                    label="Редактировать"
                    icon="pi pi-pencil"
                    onClick={() => window.location.href = `/sites/${id}/edit`}
                />
            </div>
            <Card title={site.name} subTitle={`ID: ${site.id}`}>
                <div className="flex flex-column gap-3">
                    <p>
                        <strong>URL:</strong>{' '}
                        <a href={site.url} target="_blank" rel="noopener noreferrer">
                            {site.url}
                        </a>
                    </p>
                    <p>
                        <strong>Порт:</strong> {site.port}
                    </p>
                    <div>
                        <strong>Nginx конфигурация (на диске):</strong>
                        <Button
                            label={showNginx ? 'Скрыть' : 'Показать'}
                            icon={showNginx ? 'pi pi-chevron-up' : 'pi pi-chevron-down'}
                            onClick={() => setShowNginx(!showNginx)}
                            className="p-button-text p-button-sm mt-2"
                        />
                        {showNginx && (
                            <pre style={{
                                backgroundColor: 'var(--surface-ground)',
                                padding: '16px',
                                borderRadius: '6px',
                                marginTop: '8px',
                                overflow: 'auto',
                                maxHeight: '400px',
                                fontFamily: 'monospace',
                                fontSize: '13px',
                                border: '1px solid var(--surface-border)',
                            }}>
                                {nginxData?.nginx_config || '(пусто)'}
                            </pre>
                        )}
                    </div>
                </div>
            </Card>
        </div>
    )
}
