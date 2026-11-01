import { useParams } from 'react-router-dom'
import { useGetServiceByIdQuery } from '../../services/servicesApi'
import { Card } from 'primereact/card'

export default function ServiceDetails() {
    const { id } = useParams<{ id: string }>()
    const serviceId = Number(id)

    const { data: service, isLoading, error } = useGetServiceByIdQuery(serviceId)

    if (isLoading) return <div>Загрузка...</div>
    if (error || !service) return <div>Сервис не найден</div>

    return (
        <div style={{ padding: '20px', maxWidth: '800px', margin: '0 auto' }}>
            <Card title={service.name} subTitle={`ID: ${service.id}`}>
                <p>
                    <strong>Сайт ID:</strong> {service.site_id}
                </p>
                <p>
                    <strong>Порт:</strong> {service.port}
                </p>
                <p>
                    <strong>Рабочая директория:</strong> {service.working_dir || '-'}
                </p>
                <p>
                    <strong>Команда запуска:</strong> <code>{service.startup_cmd}</code>
                </p>
                {service.comment && (
                    <p>
                        <strong>Комментарий:</strong> {service.comment}
                    </p>
                )}
                {service.env_content && (
                    <p>
                        <strong>.env файл:</strong>
                        <pre style={{ background: '#f5f5f5', padding: '10px', borderRadius: '4px', marginTop: '5px' }}>
                            {service.env_content}
                        </pre>
                    </p>
                )}
            </Card>
        </div>
    )
}
