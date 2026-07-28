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
            <Card title={service.title} subTitle={`ID: ${service.id}`}>
                <p>
                    <strong>Описание:</strong> {service.description}
                </p>
                <p>
                    <strong>Статус:</strong>{' '}
                    <span style={{ 
                        color: service.status === 'active' ? 'green' : 
                               service.status === 'inactive' ? 'red' : 'orange' 
                    }}>
                        {service.status === 'active' ? 'Активный' : 
                         service.status === 'inactive' ? 'Неактивный' : 'Ожидает'}
                    </span>
                </p>
            </Card>
        </div>
    )
}
