import { useEffect, useRef, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Toast } from 'primereact/toast'
import { SiteForm, type SiteFormValues } from '../../components/SiteForm/SiteForm'
import { useGetSiteByIdQuery, useUpdateSiteMutation, useUpdateNginxConfigMutation, useGetNginxConfigQuery, useGetSiteServicesQuery, useUpdateSiteServicesMutation } from '../../services/sitesApi'

export default function EditSite() {
  const { id } = useParams<{ id: string }>()
  const siteId = Number(id)
  const navigate = useNavigate()
  const toast = useRef<Toast>(null)
  const [updateSite, { isLoading: isUpdating, isSuccess, isError, error }] = useUpdateSiteMutation()
  const [updateNginxConfig, { isLoading: isUpdatingNginx }] = useUpdateNginxConfigMutation()
  const [updateSiteServices, { isLoading: isUpdatingServices }] = useUpdateSiteServicesMutation()

  const { data: site, isLoading: isLoadingSite } = useGetSiteByIdQuery(siteId)
  const { data: nginxData, isLoading: isLoadingNginx } = useGetNginxConfigQuery(siteId, { skip: !siteId })
  const { data: siteServices } = useGetSiteServicesQuery(siteId, { skip: !siteId })

  const [nginxConfig, setNginxConfig] = useState('')

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

  useEffect(() => {
    if (nginxData) {
      setNginxConfig(nginxData.nginx_config)
    }
  }, [nginxData])

  const handleNginxConfigSave = () => {
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
      // Save service paths to temp storage for form to pick up
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      ;(window as any).__tempServicePaths = values.services

      updateSite({
        ...site,
        name: values.name,
        url: values.url,
        port: values.port,
        status: values.status as 'active' | 'inactive' | 'pending',
      })
      .unwrap()
      .then(() => {
        // Update site services
        if (values.services && values.services.length > 0) {
          updateSiteServices({
            id: site.id,
            services: values.services,
          }).then(() => {
            showToast('success', 'Сервисы и nginx конфигурация обновлены')
          }).catch((err) => {
            showToast('error', err?.message || 'Ошибка при обновлении сервисов')
          })
        }
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
        loading={isUpdating || isUpdatingServices}
        initialValues={{
          name: site.name,
          url: site.url,
          port: site.port,
          status: site.status,
          services: siteServices?.map((ss: any) => ({
            service_id: ss.service_id,
            path: ss.path,
          })) || [],
        }}
        isEdit
      />

      <div className="max-w-2xl mx-auto">
        <div className="flex justify-content-between align-items-center mb-2">
          <h3 className="m-0">Nginx конфигурация (файл на диске)</h3>
          <button
            className="p-button p-button-secondary p-button-sm"
            onClick={handleNginxConfigSave}
            disabled={isUpdatingNginx || isLoadingNginx}
          >
            {isUpdatingNginx ? 'Сохранение...' : 'Сохранить конфигурацию'}
          </button>
        </div>
        <textarea
          value={nginxConfig}
          onChange={(e) => setNginxConfig(e.target.value)}
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
