# Документация модуля `network/http`

Модуль для выполнения HTTP-запросов, построенный на строгих принципах Clean Architecture и Inversion of Control (IoC). Обеспечивает 100% изоляцию бизнес-логики от транспортного слоя: типы и зависимости сторонней библиотеки (Axios) полностью скрыты внутри интеграции.

Модуль обеспечивает строгую типизацию, полную иммутабельность объектов запросов и мощную систему расширения через перехватчики (interceptors) и middleware.

---

## 📐 Архитектура

Модуль разделен на три изолированных слоя:

1.  **Contract Layer (Контракты)**
    Чистые TypeScript-интерфейсы и типы (`IRequest`, `IRequester`, `IClientResponse`, `IHttpRequesterConfig`). Все поля запроса помечены как `readonly`. Конфигурация описывается абстрактными интерфейсами без привязки к Axios.
2.  **Implementation Layer (Реализация)**
    Базовые реализации. Включает DTO запроса (`Request`) со встроенным механизмом безопасного клонирования, фабрику для его создания и независимую иерархию исключений.
3.  **Integration Layer (Интеграция)**
    Адаптеры для конкретных библиотек (Axios) и UI-фреймворков (React). Интеграция с React реализована по паттерну IoC: провайдер не знает о создании Axios-инстанса, он принимает готовый абстрактный `IRequester`.

---

## 🚀 Быстрый старт

### Ванильный TypeScript

Для создания клиента используется фабрика. Обратите внимание: вы передаете абстрактный `IHttpRequesterConfig`, модуль сам сконфигурирует Axios под капотом.

```typescript
import { HttpRequesterFactory, type IClientResponse } from './http';

const api = HttpRequesterFactory.create({
  baseURL: 'https://jsonplaceholder.typicode.com',
  timeout: 5000,
});

interface Todo {
  id: number;
  title: string;
  completed: boolean;
}

async function fetchTodos() {
  try {
    const response: IClientResponse<Todo[]> = await api.get<Todo[]>('/todos');
    console.log(response.data); 
  } catch (error: unknown) {
    console.error('Ошибка запроса:', error);
  }
}
```

### Интеграция с React (IoC подход)

В React мы используем инверсию управления. Сначала мы **создаем** сконфигурированный экземпляр Requester-а на уровне корня приложения, а затем **передаем его** в провайдер. Провайдеру не нужны настройки Axios, ему нужен только готовый контракт.

```tsx
import { HttpRequesterFactory, HttpRequesterProvider, useHttpRequester } from './http';

// 1. Инфраструктурный слой: создаем и конфигурируем клиент
const apiRequester = HttpRequesterFactory.create({
  baseURL: 'https://api.example.com',
});

function App() {
  return (
    // 2. Провайдер принимает ГОТОВЫЙ requesterInstance
    <HttpRequesterProvider requesterInstance={apiRequester}>
      <UserProfile />
    </HttpRequesterProvider>
  );
}

function UserProfile() {
  // 3. Хук достает абстракцию из контекста
  const requester = useHttpRequester();

  useEffect(() => {
    requester.get('/user/me').then(console.log);
  }, [requester]);

  return <div>Профиль пользователя</div>;
}
```

---

## ⚙️ Расширенное использование

### Перехватчики (Interceptors) и Иммутабельность

Поскольку объект запроса (`IRequest`) полностью иммутабелен (`readonly`), вы не можете напрямую модифицировать его свойства. 

Для модификации запроса в перехватчике используется метод **`request.clone()`**, который делает частичное (shallow merge) копирование объекта, сохраняя оригинал в неприкосновенности.

```typescript
import { HttpRequest, HttpRequesterFactory } from './http';

const requester = HttpRequesterFactory.create();

requester.addRequestInterceptor((request) => {
  const token = localStorage.getItem('auth_token');

  // Безопасное клонирование с переопределением заголовков
  if (token && request instanceof HttpRequest) {
    return request.clone({ 
      headers: { ...request.headers, Authorization: `Bearer ${token}` } 
    });
  }

  // Фоллбэк, если используется объект, не являющийся экземпляром HttpRequest
  return { 
    ...request, 
    headers: { ...request.headers, Authorization: `Bearer ${token}` } 
  };
});
```

### Middleware (Промежуточные обработчики)

Middleware строится в виде цепочки ответственности через функцию `next`. Подходит для реализации сложной логики (кэширование, retries), где нужно полностью контролировать вызов следующего этапа.

```typescript
import { HttpRequesterFactory, type IHttpRequesterMiddleware } from './http';

const retryMiddleware: IHttpRequesterMiddleware = async (request, next) => {
  try {
    return await next(request);
  } catch (error) {
    // Логика повтора запроса
    if (someCondition) return await next(request);
    throw error;
  }
};

const requester = HttpRequesterFactory.create();
requester.use(retryMiddleware);
```

### Создание производных экземпляров

Метод `create` позволяет создать новый `Requester`, наследующий базовые настройки текущего (удобно для микросервисов). Перехватчики и middleware главного API **не копируются** в новый экземпляр.

```typescript
import { HttpRequesterFactory } from './http';

const mainApi = HttpRequesterFactory.create({ baseURL: 'https://api.mysite.com/v1' });

// Создаем изолированную копию с новым baseURL
const authApi = mainApi.create({ baseURL: 'https://auth.mysite.com/v1' });
```

---

## 🛡️ Обработка ошибок

Модуль использует "плоскую" (независимую) иерархию исключений. Все ошибки реализуют общий базовый контракт `IHttpException`, но при этом **не наследуются друг от друга**. Это исключает пересечение зон ответственности и делает проверки через `instanceof` предсказуемыми и строгими.

### Зоны ответственности исключений:
1.  **`HttpClientException`** — Ошибка транспортного уровня. Выбрасывается низкоуровневым адаптером, если Axios не смог выполнить запрос (отвал сети, таймаут, сбой парсинга, ответ сервера с ошибкой).
2.  **`HttpRequestException`** — Ошибка структуры запроса (DTO). Выбрасывается на этапе валидации или сборки объекта запроса *до* отправки в сеть (например, невалидный URL).
3.  **`HttpRequesterException`** — Ошибка пайплайна. Выбрасывается `Requester`'ом, если ошибка транспорта была перехвачена, изменена (модифицировано сообщение или статус) в `Error Interceptors` и проброшена дальше.

### Сохранение Stack Trace
Пайплайн спроектирован так, чтобы минимизировать обертки:
* Если `Error Interceptors` вернули *ту же самую* ошибку без изменений, она пробрасывается как есть.
* Если интерцептор вернул *новый объект* ошибки, пайплайн оборачивает его в `HttpRequesterException`, но сохраняет оригинал в свойство `details`.

### Пример обработки в коде:

```typescript
import { 
  HttpClientException, 
  HttpRequesterException, 
  HttpRequestException 
} from './http';

try {
  await requester.get('/protected');
} catch (error: unknown) {
  // Порядок проверок важен! Благодаря плоской иерархии они не перекрывают друг друга.

  if (error instanceof HttpRequesterException) {
    // Ошибка была изменена перехватчиками пайплайна. 
    // Оригинальная ошибка доступна в error.details
    console.error('Модифицированная ошибка пайплайна:', error.message);
  } 
  else if (error instanceof HttpClientException) {
    // Чистая ошибка транспорта/сети. 
    // Stack Trace в консоли будет указывать прямо на место падения!
    console.error(`Транспортная ошибка [${error.statusCode}]:`, error.message);
  }
  else if (error instanceof HttpRequestException) {
    // Ошибка сборки или валидации DTO запроса
    console.error(`Ошибка запроса [${error.statusCode}]:`, error.message);
  }
}
```

---

## 📚 Справочник API (Экспорты)

### Базовые типы (Contracts)
*   `IHttpException` — Базовый интерфейс для всех исключений модуля (`message`, `statusCode`, `details?`).
*   `IRequest<TData>` — Структура исходящего запроса (строго `readonly`).
*   `IClientResponse<T>` — Структура ответа (`{ data: T, status: number }`).
*   `IHttpRequesterConfig` — Абстрактный интерфейс конфигурации (`baseURL`, `headers`, `timeout`). Не содержит типов Axios.
*   `THttpRequestMethod` — `'get' | 'post' | 'patch' | 'put' | 'delete'`.
*   `IHttpRequesterInterceptor` — Интерфейс для перехватчиков запроса/ответа/ошибки.
*   `IHttpRequesterMiddleware` — Тип функции для middleware `(req, next) => Promise`.
*   `THttpRequesterDispatch` — Тип функции `next`.

### Классы (Реализация)
*   `HttpRequest` — Реализация DTO запроса.
    *   **`clone(partial: Partial<IRequest<TData>>): HttpRequest<TData>`** — Создает новый экземпляр запроса, сливая текущие данные с переданными изменениями.

### Классы (Исключения)
Все исключения независимы друг от друга (не наследуются по цепочке), но имеют идентичную структуру.
*   `HttpClientException` — Ошибка транспортного уровня (дефолтный `statusCode: 0`).
*   `HttpRequestException` — Ошибка уровня валидации DTO запроса (дефолтный `statusCode: 400`).
*   `HttpRequesterException` — Ошибка пайплайна Requester'а. Содержит исходную ошибку в `details`.

### Интеграция (Скрытая под капотом)
*   `HttpClient` — Низкоуровневый адаптер (реализует `IClient`).
*   `HttpClientFactory.create(config?)` — Фабрика клиента.
*   `HttpRequester` — Полноценный обработчик с поддержкой Interceptors/Middleware.
*   `HttpRequesterFactory.create(config?)` — Фабрика обработчика. Принимает `IHttpRequesterConfig`.

### React
*   `HttpRequesterProvider` — Провайдер контекста (IoC).
*   `IHttpRequesterProviderProps` — Пропсы провайдера (`children`, `requesterInstance: IRequester`). **Не принимает конфигурацию Axios напрямую**.
*   `useHttpRequester()` — Хук для получения экземпляра обработчика из контекста. Бросает ошибку вне провайдера.
