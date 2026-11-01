import { Formik, Form, Field, ErrorMessage, type FormikHelpers } from 'formik'
import * as Yup from 'yup'
import { Card } from 'primereact/card'
import { InputText } from 'primereact/inputtext'
import { Dropdown } from 'primereact/dropdown'
import { InputTextarea } from 'primereact/inputtextarea'
import { Button } from 'primereact/button'
import { useGetSitesQuery } from '../../services/sitesApi'

export type ServiceFormValues = {
  site_id: number
  name: string
  port: number
  working_dir: string
  startup_cmd: string
  env_content: string
  comment: string
  path: string
}

type ServiceFormProps = {
  onSubmit: (values: ServiceFormValues & { path?: string }) => void
  onCancel: () => void
  loading?: boolean
  initialSiteId?: number
}

const validationSchema = Yup.object({
  site_id: Yup.number()
    .required('Выберите сайт')
    .min(1, 'Сайт обязателен'),
  name: Yup.string()
    .required('Название обязательно')
    .min(1, 'Название обязательно'),
  port: Yup.number()
    .required('Порт обязателен')
    .min(1, 'Порт должен быть больше 0')
    .max(65535, 'Порт должен быть до 65535'),
  working_dir: Yup.string()
    .optional(),
  startup_cmd: Yup.string()
    .required('Команда запуска обязательна'),
  env_content: Yup.string()
    .optional(),
  comment: Yup.string()
    .optional(),
  path: Yup.string()
    .optional(),
})

function FormField({
  name,
  label,
  type = 'text',
}: {
  name: keyof ServiceFormValues
  label: string
  type?: string
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
            type={type}
            value={String(field[name])}
            onChange={(e: any) => {
              const val = type === 'number' ? parseInt(e.target.value, 10) : e.target.value
              form.setFieldValue(name, val)
            }}
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
  options: { label: string; value: number }[]
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
            value={field[name] as number}
            options={options}
            optionLabel="label"
            optionValue="value"
            onChange={(e: any) => form.setFieldValue(name, e.value)}
            className="w-full"
            placeholder="Выберите..."
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

function FormTextArea({
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
          <InputTextarea
            id={name}
            value={field[name] as string}
            onChange={(e: any) => form.setFieldValue(name, e.target.value)}
            className="w-full"
            rows={4}
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

export function ServiceForm({ onSubmit, onCancel, loading = false, initialSiteId }: ServiceFormProps) {
  const handleSubmit = (values: ServiceFormValues & { path?: string }, helpers: FormikHelpers<ServiceFormValues & { path?: string }>) => {
    const formattedValues: ServiceFormValues = {
      ...values,
      path: values.path || '/',
    }
    onSubmit(formattedValues)
    helpers.setSubmitting(false)
  }

  const { data: sites } = useGetSitesQuery()

  const siteOptions = sites?.map((site) => ({
    label: site.name,
    value: site.id,
  })) || []

  return (
    <Card title="Создание сервиса" className="max-w-2xl mx-auto">
      <Formik
        initialValues={{
          site_id: initialSiteId || 0,
          name: '',
          port: 3000,
          working_dir: '',
          startup_cmd: '',
          env_content: '',
          comment: '',
          path: '/',
        }}
        validationSchema={validationSchema}
        onSubmit={handleSubmit}
      >
        {({ isSubmitting, dirty, isValid }) => (
          <Form className="p-fluid">
            <FormDropdown name="site_id" label="Сайт" options={siteOptions} />
            <FormField name="name" label="Название сервиса" />
            <FormField name="port" label="Порт" type="number" />
            <FormField name="working_dir" label="Рабочая директория" />
            <FormField name="startup_cmd" label="Команда запуска" />
            <FormTextArea name="env_content" label="Содержимое .env файла" />
            <FormField name="path" label="Путь для nginx (например: /api)" />
            <FormField name="comment" label="Комментарий" />

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
