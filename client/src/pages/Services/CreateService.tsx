import { useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { Toast } from 'primereact/toast'
import { ServiceForm, type ServiceFormValues } from '../../components/ServiceForm/ServiceForm'
import { useAddServiceMutation } from '../../services/servicesApi'

export default function CreateService() {
  const navigate = useNavigate()
  const [addService, { isLoading, isSuccess, isError, error }] = useAddServiceMutation()
  const toast = useRef<Toast>(null)

  const showToast = (severity: 'success' | 'error', message: string) => {
    if (toast.current) {
      toast.current.show({
        severity,
        summary: severity === 'success' ? 'Успех' : 'Ошибка',
        detail: message,
        life: 3000,
      })
    }
  }

  useEffect(() => {
    if (isSuccess) {
      showToast('success', 'Сервис успешно создан')
      navigate('/services')
    }
    if (isError && error) {
      const errorMessage = 'data' in error && error.data && typeof error.data === 'object' && 'message' in error.data
        ? (error.data as { message?: string }).message
        : 'Ошибка при создании сервиса'
      if (errorMessage) {
        showToast('error', errorMessage)
      }
    }
  }, [isSuccess, isError, error, navigate])

  const handleSuccess = (values: ServiceFormValues) => {
    addService(values)
  }

  return (
    <div className="p-4">
      <ServiceForm
        onSubmit={handleSuccess}
        onCancel={() => navigate('/services')}
        loading={isLoading}
      />
      <Toast ref={toast} />
    </div>
  )
}
