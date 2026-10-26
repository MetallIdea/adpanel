import { Formik, Form, Field, ErrorMessage, type FormikHelpers } from 'formik'
import * as Yup from 'yup'
import { Card } from 'primereact/card'
import { InputText } from 'primereact/inputtext'
import { Button } from 'primereact/button'
import { Dropdown } from 'primereact/dropdown'
import { useCallback } from 'react'
import { useGetServicesQuery } from '../../services/servicesApi'

export type SiteFormValues = {
  name: string
  url: string
  port: number
  status: string
  services: ServicePath[]
}

export type ServicePath = {
  service_id: number
  path: string
}

type SiteFormProps = {
  onSubmit: (values: SiteFormValues) => void
  onCancel: () => void
  loading?: boolean
  initialValues?: Partial<SiteFormValues>
  isEdit?: boolean
}

const validationSchema = Yup.object({
  name: Yup.string()
    .required('Название обязательно')
    .min(3, 'Минимум 3 символа'),
  url: Yup.string()
    .nullable()
    .url('Неверный формат URL'),
  port: Yup.number()
    .required('Порт обязателен')
    .min(1, 'Минимум 1')
    .max(65535, 'Максимум 65535'),
  status: Yup.string()
    .required('Статус обязателен'),
})

function FormField({
  name,
  label,
  type = 'text',
  placeholder,
}: {
  name: keyof SiteFormValues
  label: string
  type?: string
  placeholder?: string
}) {
  return (
    <Field name={name}>
      {({ field, form }: { field: any; form: any }) => {
        return (
          <div className="p-field mb-4">
            <label htmlFor={name} className="p-label">
              {label}
            </label>
            <InputText
              id={name}
              type={type}
              value={field.value}
              onChange={(e: any) => form.setFieldValue(name, type === 'number' ? Number(e.target.value) : e.target.value)}
              placeholder={placeholder}
              className="w-full"
            />
            <ErrorMessage
              name={name}
              component={() => (
                <small className="p-error block mt-1">{form.errors[name]}</small>
              )}
            />
          </div>
        )
      }}
    </Field>
  )
}

export function SiteForm({ onSubmit, onCancel, loading = false, initialValues, isEdit = false }: SiteFormProps) {
  const handleSubmit = (values: SiteFormValues, helpers: FormikHelpers<SiteFormValues>) => {
    onSubmit(values)
    helpers.setSubmitting(false)
  }

  const { data: services } = useGetServicesQuery()

  const serviceOptions = services?.map((s: any) => ({
    label: s.name,
    value: s.id,
  })) || []

  const addServicePath = useCallback(() => {
    const currentServices = initialValues?.services || []
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    ;(window as any).__tempServicePaths = [
      ...currentServices,
      { service_id: 0, path: '/' }
    ]
  }, [initialValues])

  const updateServicePath = useCallback((index: number, field: string, value: any) => {
    const currentServices = initialValues?.services || []
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    ;(window as any).__tempServicePaths = currentServices.map((s: any, i: number) => 
      i === index ? { ...s, [field]: value } : s
    )
  }, [initialValues])

  const removeServicePath = useCallback((index: number) => {
    const currentServices = initialValues?.services || []
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    ;(window as any).__tempServicePaths = currentServices.filter((_: any, i: number) => i !== index)
  }, [initialValues])

  return (
    <Card title={isEdit ? 'Редактирование сайта' : 'Создание сайта'} className="max-w-2xl mx-auto">
      <Formik
        initialValues={{
          name: initialValues?.name || '',
          url: initialValues?.url || '',
          port: initialValues?.port || 8080,
          status: initialValues?.status || 'active',
          services: initialValues?.services || [],
        }}
        validationSchema={validationSchema}
        onSubmit={handleSubmit}
      >
        {({ isSubmitting, dirty, isValid, values, setFieldValue }) => {
          // Sync temp service paths with form
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const tempPaths = (window as any).__tempServicePaths
          if (tempPaths && JSON.stringify(tempPaths) !== JSON.stringify(values.services)) {
            setFieldValue('services', tempPaths)
          }

          return (
            <Form className="p-fluid">
              <FormField name="name" label="Название" />
              <FormField name="url" label="URL" placeholder="https://example.com" />
              <FormField name="port" label="Порт" type="number" />
              <FormField name="status" label="Статус" type="text" />

              <div className="p-field mb-4">
                <label className="p-label">Сервисы и пути</label>
                <small className="text-secondary block mb-2">
                  Укажите сервисы, которые будут доступны по указанным путям через nginx
                </small>
                
                {values.services.map((sp, index) => (
                  <div key={index} className="flex gap-2 mb-2 align-items-center">
                    <Dropdown
                      value={sp.service_id}
                      options={serviceOptions}
                      optionLabel="label"
                      optionValue="value"
                      onChange={(e: any) => updateServicePath(index, 'service_id', e.value)}
                      placeholder="Выберите сервис"
                      className="flex-1"
                    />
                    <InputText
                      value={sp.path}
                      onChange={(e: any) => updateServicePath(index, 'path', e.target.value)}
                      placeholder="/path"
                      className="flex-1"
                    />
                    <Button
                      type="button"
                      icon="pi pi-trash"
                      onClick={() => removeServicePath(index)}
                      className="p-button-danger p-button-sm"
                      style={{ width: '38px', minWidth: '38px' }}
                    />
                  </div>
                ))}
                
                <Button
                  type="button"
                  label="Добавить сервис"
                  icon="pi pi-plus"
                  onClick={addServicePath}
                  className="p-button-sm p-button-outlined mt-2"
                />
              </div>

              <div className="flex gap-2 mt-6">
                <Button
                  type="submit"
                  label={isEdit ? 'Сохранить' : 'Создать'}
                  icon="pi pi-check"
                  disabled={isSubmitting || !dirty || !isValid}
                  loading={loading}
                  className="flex-1"
                />
                <Button
                  type="button"
                  label="Отмена"
                  icon="pi pi-times"
                  onClick={onCancel}
                  className="flex-1 p-button-secondary"
                />
              </div>
            </Form>
          )
        }}
      </Formik>
    </Card>
  )
}
