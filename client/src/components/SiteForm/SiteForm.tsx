import { Formik, Form, Field, ErrorMessage, type FormikHelpers } from 'formik'
import * as Yup from 'yup'
import { Card } from 'primereact/card'
import { InputText } from 'primereact/inputtext'
import { Button } from 'primereact/button'

export type SiteFormValues = {
  name: string
  url: string
  port: number
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
        console.log(field);
        return (
          <div className="p-field mb-4">
            <label htmlFor={name} className="p-label">
              {label}
            </label>
            <InputText
              id={name}
              type={type}
              value={field.value}
              onChange={(e) => form.setFieldValue(name, type === 'number' ? Number(e.target.value) : e.target.value)}
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

  console.log(initialValues);

  return (
    <Card title={isEdit ? 'Редактирование сайта' : 'Создание сайта'} className="max-w-2xl mx-auto">
      <Formik
        initialValues={{
          name: initialValues?.name || '',
          url: initialValues?.url || '',
          port: initialValues?.port || 8080,
        }}
        validationSchema={validationSchema}
        onSubmit={handleSubmit}
      >
        {({ isSubmitting, dirty, isValid }) => (
          <Form className="p-fluid">
            <FormField name="name" label="Название" />
            <FormField name="url" label="URL" placeholder="https://example.com" />
            <FormField name="port" label="Порт" type="number" />

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
        )}
      </Formik>
    </Card>
  )
}
