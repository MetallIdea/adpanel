import { Formik, Form, Field, ErrorMessage } from 'formik'
import * as Yup from 'yup'
import { Card } from 'primereact/card'
import { InputText } from 'primereact/inputtext'
import { Dropdown } from 'primereact/dropdown'
import { Button } from 'primereact/button'

export type ServiceFormValues = {
  title: string
  description: string
  status: 'active' | 'inactive' | 'pending'
}

type ServiceFormProps = {
  onSubmit: (values: ServiceFormValues) => void
  onCancel: () => void
  loading?: boolean
}

const statusOptions = [
  { label: 'Активный', value: 'active' },
  { label: 'Неактивный', value: 'inactive' },
  { label: 'Ожидает', value: 'pending' },
]

const validationSchema = Yup.object({
  title: Yup.string()
    .required('Название обязательно')
    .min(3, 'Минимум 3 символа'),
  description: Yup.string()
    .optional(),
  status: Yup.string()
    .required('Статус обязателен')
    .oneOf(['active', 'inactive', 'pending']),
})

export function ServiceForm({ onSubmit, onCancel, loading = false }: ServiceFormProps) {
  const handleSubmit = (values: ServiceFormValues, { setSubmitting }: { setSubmitting: (isSubmitting: boolean) => void }) => {
    onSubmit(values)
    setSubmitting(false)
  }

  return (
    <Card title="Создание сервиса" className="max-w-2xl mx-auto">
      <Formik
        initialValues={{
          title: '',
          description: '',
          status: 'active',
        }}
        validationSchema={validationSchema}
        onSubmit={handleSubmit}
      >
        {({ isSubmitting, dirty, isValid }) => (
          <Form className="p-fluid">
            <div className="p-field mb-4">
              <label htmlFor="title" className="p-label">
                Название
              </label>
              <Field
                id="title"
                name="title"
                component={InputText}
                className="w-full"
              />
              <ErrorMessage
                name="title"
                component={() => (
                  <small className="p-error block mt-1">Название обязательно</small>
                )}
              />
            </div>

            <div className="p-field mb-4">
              <label htmlFor="description" className="p-label">
                Описание
              </label>
              <Field
                id="description"
                name="description"
                component={InputText}
                className="w-full"
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
