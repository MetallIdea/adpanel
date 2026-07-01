import { useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { Toast } from 'primereact/toast'
import { SiteForm, type SiteFormValues } from '../../components/Sites/SiteForm'
import { useAddSiteMutation } from '../../services/sitesApi'

export default function CreateSite() {
  const navigate = useNavigate()
  const [addSite, { isLoading, isSuccess, isError, error }] = useAddSiteMutation()
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
      showToast('success', 'Сайт успешно создан')
      navigate('/sites')
    }
    if (isError && error) {
      const errorMessage = 'data' in error && error.data && typeof error.data === 'object' && 'message' in error.data
        ? (error.data as { message?: string }).message
        : 'Ошибка при создании сайта'
      if (errorMessage) {
        showToast('error', errorMessage)
      }
    }
  }, [isSuccess, isError, error, navigate])

  const handleSuccess = (values: SiteFormValues) => {
    addSite(values)
  }

  return (
    <div className="p-4">
      <SiteForm
        onSubmit={handleSuccess}
        onCancel={() => navigate('/sites')}
        loading={isLoading}
      />
      <Toast ref={toast} />
    </div>
  )
}
