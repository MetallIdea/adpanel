# План создания модуля Сервисы

## Обзор
Создать модуль Сервисы по аналогии с модулем Сайты, с обновленными требованиями:
- Поля сервиса: `id`, `title`, `description`, `status`
- API endpoint: `/api/services`

## Файлы для создания

### 1. client/src/store/servicesSlice.ts
- Тип `Service`: id (number), title (string), description (string), status ('active' | 'inactive' | 'pending')
- Тип `ServicesState`: {}

### 2. client/src/services/servicesApi.ts
- `getServices`: query для получения всех сервисов
- `getServiceById`: query для получения сервиса по ID
- `addService`: mutation для создания нового сервиса

### 3. client/src/components/Services/ServicesList.tsx
- Компонент списка сервисов с PrimeReact DataTable
- Отображение: ID, title, description, status, visits

### 4. client/src/components/Services/ServiceForm.tsx
- Форма создания сервиса с Formik + Yup
- Поля: title, description, status
- Валидация: title (обязательно, минимум 3 символа), description (необязательно), status (обязательный)

### 5. client/src/pages/Services/Services.tsx
- Страница списка сервисов
- Кнопка "Создать сервис"

### 6. client/src/pages/Services/CreateService.tsx
- Страница создания сервиса
- Использует ServiceForm
- Обработка успеха/ошибки с Toast

### 7. client/src/pages/Services/ServiceDetails.tsx
- Страница деталей сервиса
- Отображение: title, description, url, status

### 8. client/src/App.tsx
- Добавить роуты:
  - `/services` -> Services
  - `/services/create` -> CreateService
  - `/services/:id` -> ServiceDetails

### 9. client/src/store/store.ts
- Добавить servicesReducer
- Добавить servicesApi.reducerPath
- Добавить servicesApi.middleware

## Шаги реализации
1. Создать типы и API слой
2. Создать компоненты UI
3. Создать страницы
4. Обновить маршрутизацию и хранилище
5. Протестировать функциональность
