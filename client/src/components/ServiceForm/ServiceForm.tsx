import { Formik, Form, Field, ErrorMessage, type FormikHelpers } from 'formik'
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

function FormField({
  name,
  label,
}: {
  name: keyof ServiceFormValues
  label: string
}) {
  return (
    <Field name={name}>
      {({ field, form }: { field: any; form: any }) => (
        <div className="p-field mb-4">
          <label htmlFor={name} className="p-label">
            {label}
          </label>
          <InputText
            id={name}
            value={field[name] as string}
            onChange={(e) => form.setFieldValue(name, e.target.value)}
            className="w-full"
          />
          <ErrorMessage
            name={name}
            component={() => (
              <small className="p-error block mt-1">{form.errors[name]}</small>
            )}
          />
        </div>
      )}
    </Field>
  )
}

function FormDropdown({
  name,
  label,
  options,
}: {
  name: keyof ServiceFormValues
  label: string
  options: { label: string; value: string }[]
}) {
  return (
    <Field name={name}>
      {({ field, form }: { field: any; form: any }) => (
        <div className="p-field mb-4">
          <label htmlFor={name} className="p-label">
            {label}
          </label>
          <Dropdown
            id={name}
            value={field[name] as string}
            options={options}
            optionLabel="label"
            optionValue="value"
            onChange={(e) => form.setFieldValue(name, e.value)}
            className="w-full"
          />
          <ErrorMessage
            name={name}
            component={() => (
              <small className="p-error block mt-1">{form.errors[name]}</small>
            )}
          />
        </div>
      )}
    </Field>
  )
}

export function ServiceForm({ onSubmit, onCancel, loading = false }: ServiceFormProps) {
  const handleSubmit = (values: ServiceFormValues, helpers: FormikHelpers<ServiceFormValues>) => {
    onSubmit(values)
    helpers.setSubmitting(false)
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
            <FormField name="title" label="Название" />
            <FormField name="description" label="Описание" />
            <FormDropdown name="status" label="Статус" options={statusOptions} />

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
