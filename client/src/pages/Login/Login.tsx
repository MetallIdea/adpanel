import { Formik, Form, Field, type FormikHelpers } from 'formik'
import * as yup from 'yup'
import { useState } from 'react'
import { login } from '../../services/authApi'
import styles from './Login.module.css'

interface LoginFormValues {
  email: string
  password: string
}

const validationSchema = yup.object({
  email: yup
    .string()
    .email('Введите корректный email')
    .required('Email обязателен'),
  password: yup
    .string()
    .min(6, 'Пароль должен содержать минимум 6 символов')
    .required('Пароль обязателен'),
})

export default function Login() {
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (
    values: LoginFormValues,
    helpers: FormikHelpers<LoginFormValues>
  ) => {
    setError(null)
    setLoading(true)

    try {
      await login(values)
      // TODO: redirect after successful login
      helpers.resetForm()
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Произошла неизвестная ошибка')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.wrapper}>
      <div className={styles.card}>
        <h1 className={styles.title}>Вход в систему</h1>

        <Formik
          initialValues={{ email: '', password: '' }}
          validationSchema={validationSchema}
          onSubmit={handleSubmit}
        >
          {({ errors, touched }) => (
            <Form className={styles.form}>
              {error && <div className={styles.errorText}>{error}</div>}

              <div className={styles.field}>
                <label htmlFor="email" className={styles.label}>
                  Email
                </label>
                <Field
                  id="email"
                  name="email"
                  type="email"
                  className={`${styles.input} ${errors.email && touched.email ? styles.inputError : ''
                    }`}
                  placeholder="example@mail.com"
                />
                {errors.email && touched.email && (
                  <span className={styles.error}>{errors.email}</span>
                )}
              </div>

              <div className={styles.field}>
                <label htmlFor="password" className={styles.label}>
                  Пароль
                </label>
                <Field
                  id="password"
                  name="password"
                  type="password"
                  className={`${styles.input} ${errors.password && touched.password
                      ? styles.inputError
                      : ''
                    }`}
                  placeholder="••••••••"
                />
                {errors.password && touched.password && (
                  <span className={styles.error}>{errors.password}</span>
                )}
              </div>

              <button
                type="submit"
                className={styles.button}
                disabled={loading}
              >
                {loading ? 'Вход...' : 'Войти'}
              </button>
            </Form>
          )}
        </Formik>
      </div>
    </div>
  )
}
