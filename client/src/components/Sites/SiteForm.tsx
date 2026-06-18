import { Formik, Form, Field, ErrorMessage } from 'formik'
import * as Yup from 'yup'
import { Card } from 'primereact/card'
import { InputText } from 'primereact/inputtext'
import { Dropdown } from 'primereact/dropdown'
import { Button } from 'primereact/button'

export interface SiteFormValues {
  name: string
  url: string
  status: 'active' | 'inactive' | 'pending'
}

interface SiteFormProps {
  onSubmit: (values: SiteFormValues) => void
  onCancel: () => void
  loading?: boolean
}

const statusOptions = [
  { label: 'Активный', value: 'active' },
  { label: 'Неактивный', value: 'inactive' },
  { label: 'Ожидает', value: 'pending' },
]

const validationSchema = Yup.object({
  name: Yup.string()
    .required('Название обязательно')
    .min(3, 'Минимум 3 символа'),
  url: Yup.string()
    .required('URL обязателен')
    .url('Неверный формат URL'),
  status: Yup.string()
    .required('Статус обязателен')
    .oneOf(['active', 'inactive', 'pending']),
})

export function SiteForm({ onSubmit, onCancel, loading = false }: SiteFormProps) {
  const handleSubmit = (values: SiteFormValues, { setSubmitting }: { setSubmitting: (isSubmitting: boolean) => void }) => {
    onSubmit(values)
    setSubmitting(false)
  }

  return (
    <Card title="Создание сайта" className="max-w-2xl mx-auto">
      <Formik
        initialValues={{
          name: '',
          url: '',
          status: 'active',
        }}
        validationSchema={validationSchema}
        onSubmit={handleSubmit}
      >
        {({ isSubmitting, dirty, isValid }) => (
          <Form className="p-fluid">
            <div className="p-field mb-4">
              <label htmlFor="name" className="p-label">
                Название
              </label>
              <Field
                id="name"
                name="name"
                component={InputText}
                className="w-full"
              />
              <ErrorMessage
                name="name"
                component={() => (
                  <small className="p-error block mt-1">Название обязательно</small>
                )}
              />
            </div>

            <div className="p-field mb-4">
              <label htmlFor="url" className="p-label">
                URL
              </label>
              <Field
                id="url"
                name="url"
                component={InputText}
                className="w-full"
                placeholder="https://example.com"
              />
              <ErrorMessage
                name="url"
                component={() => (
                  <small className="p-error block mt-1">Неверный формат URL</small>
                )}
              />
            </div>

            <div className="p-field mb-4">
              <label htmlFor="status" className="p-label">
                Статус
              </label>
              <Field
                id="status"
                name="status"
                component={Dropdown}
                options={statusOptions}
                optionLabel="label"
                optionValue="value"
                className="w-full"
              />
              <ErrorMessage
                name="status"
                component={() => (
                  <small className="p-error block mt-1">Статус обязателен</small>
                )}
              />
            </div>

            <div className="flex gap-2 mt-6">
              <Button
                type="submit"
                label="Создать"
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
        )}
      </Formik>
    </Card>
  )
}
