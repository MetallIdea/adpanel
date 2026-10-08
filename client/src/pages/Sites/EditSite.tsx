import { useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Toast } from 'primereact/toast'
import { SiteForm, type SiteFormValues } from '../../components/SiteForm/SiteForm'
import { useGetSiteByIdQuery, useUpdateSiteMutation, useUpdateNginxConfigMutation, useGetNginxConfigQuery } from '../../services/sitesApi'

export default function EditSite() {
  const { id } = useParams<{ id: string }>()
  const siteId = Number(id)
  const navigate = useNavigate()
  const toast = useRef<Toast>(null)
  const [updateSite, { isLoading: isUpdating, isSuccess, isError, error }] = useUpdateSiteMutation()
  const [updateNginxConfig, { isLoading: isUpdatingNginx }] = useUpdateNginxConfigMutation()

  const { data: site, isLoading: isLoadingSite } = useGetSiteByIdQuery(siteId)
  const { data: nginxData, isLoading: isLoadingNginx } = useGetNginxConfigQuery(siteId, { skip: !siteId })

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
      showToast('success', 'Сайт успешно обновлён')
      navigate('/sites')
    }
    if (isError && error) {
      const errorMessage = 'data' in error && error.data && typeof error.data === 'object' && 'message' in error.data
        ? (error.data as { message?: string }).message
        : 'Ошибка при обновлении сайта'
      if (errorMessage) {
        showToast('error', errorMessage)
      }
    }
  }, [isSuccess, isError, error, navigate])

  const handleNginxConfigSave = (nginxConfig: string) => {
    if (site) {
      updateNginxConfig({ id: site.id, nginx_config: nginxConfig })
        .unwrap()
        .then(() => {
          showToast('success', 'Nginx конфигурация сохранена на диск')
        })
        .catch((err) => {
          showToast('error', err?.message || 'Ошибка при сохранении конфигурации')
        })
    }
  }

  const handleSuccess = (values: SiteFormValues) => {
    if (site) {
      updateSite({
        ...site,
        ...values,
        visits: site.visits,
        status: site.status,
      })
    }
  }

  if (isLoadingSite) return <div>Загрузка...</div>
  if (!site) return <div>Сайт не найден</div>

  return (
    <div className="p-4 space-y-4">
      <SiteForm
        onSubmit={handleSuccess}
        onCancel={() => navigate('/sites')}
        loading={isUpdating}
        initialValues={{
          name: site.name,
          url: site.url,
          port: site.port,
          status: site.status,
        }}
        isEdit
      />

      <div className="max-w-2xl mx-auto">
        <div className="flex justify-content-between align-items-center mb-2">
          <h3 className="m-0">Nginx конфигурация (файл на диске)</h3>
          <button
            className="p-button p-button-secondary p-button-sm"
            onClick={() => handleNginxConfigSave(nginxData?.nginx_config || '')}
            disabled={isUpdatingNginx || isLoadingNginx}
          >
            {isUpdatingNginx ? 'Сохранение...' : 'Сохранить конфигурацию'}
          </button>
        </div>
        <textarea
          value={nginxData?.nginx_config || ''}
          onChange={(e) => {
            if (nginxData) {
              nginxData.nginx_config = e.target.value
            }
          }}
          style={{
            width: '100%',
            minHeight: '300px',
            fontFamily: 'monospace',
            fontSize: '13px',
            padding: '16px',
            borderRadius: '6px',
            border: '1px solid var(--surface-border)',
            backgroundColor: 'var(--surface-ground)',
            resize: 'vertical',
          }}
        />
        <small className="text-secondary block mt-2">
          Путь к файлу: nginx_configs/site_{site.id}.conf
        </small>
      </div>

      <Toast ref={toast} />
    </div>
  )
}
