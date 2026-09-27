# AI Context — Client App

## Технологический стек
- **React**
- **Vite**
- **TypeScript**
- **Redux toolkit**
- **Redux toolkit query**
- **Prime react**

## Структура проекта (усечённая)
client/
├── public/
├── src/
│   ├── components/        # UI-компоненты (Button, Card...)
│   ├── constants/        # UI-компоненты (Button, Card...)
│   ├── pages/             # Страницы роутинга
│   ├── services/          # API-обёртки
│   ├── store/          # Стейт
│   ├── utils/             # Вспомогательные функции
├── .env                   # Переменные окружения
└── .gitignore

## Переменные окружения (.env)
VITE_API_BASE_URL=https://api.example.com  # по умолчанию

В .env.local переопределяется для локальной разработки

## Принятая практика API-запросов
- Все запросы через `redux toolkit query`

## Общие правила написания кода
Не использовать index.ts файлы
Не использовать export default
Экспортировать все сразу `export function` или `export const`

## Правила написания компонентов
Для страниц использовать pages/, для компонентов — components/.
Если файл уже существует — изменить, а не переписать полностью (если не просили иного).
Стили выносить в css модули `*.module.css`