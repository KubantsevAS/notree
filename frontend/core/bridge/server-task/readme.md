# Документация модуля `bridge/server-task`

Модуль для управления серверным состоянием (кэшированием, синхронизацией, фоновым обновлением), построенный на строгих принципах Clean Architecture и Inversion of Control (IoC). Обеспечивает 100% изоляцию бизнес-логики от конкретной библиотеки-реализации: типы, флаги состояния и зависимости `@tanstack/react-query` полностью скрыты внутри интеграции. Прямой импорт из `@tanstack/react-query` в компонентах строго запрещен.

Модуль обеспечивает строгую типизацию контрактов, безопасное управление кэшем через урезанный API и мощную систему нормализации ошибок по паттерну "Border Guard" (Пограничный контроль).

---

## 📐 Архитектура

Модуль разделен на три изолированных слоя:

1.  **Contract Layer (Контракты)**
    Чистые TypeScript-интерфейсы, описывающие сущности бизнес-логики (`IServerTaskQueryParams`, `IServerTaskMutationResult`, `IServerTaskClientActions`). Контракты не содержат технических деталей реализации (например, нет флагов `isFetching` или `fetchStatus`).
2.  **Implementation Layer (Реализация)**
    Базовые реализации и утилиты. Включает строгую иерархию исключений (с правильным прототипным наследованием) и независимую функцию нормализации ошибок `normalizeGeneralException`, которая приводит любой `unknown` к стандартному виду `{ message, code }`.
3.  **Integration Layer (Интеграция)**
    Адаптеры для React и `@tanstack/react-query`. Реализует паттерн IoC: адаптеры не знают, как выполнять HTTP-запросы. Сетевой слой пробрасывается внутрь через функцию `executor` в момент вызова хука. Выступает "границей" модуля, перехватывая ошибки из `executor` и оборачивая их в доменные исключения.

---

## 🚀 Быстрый старт

### Инициализация (Провайдер)

Оберните корень приложения в `ServerTaskClientProvider`. Провайдер принимает абстрактный контракт `IServerTaskClientConfig` и скрывает внутри себя создание и настройку технического `QueryClient`.

```tsx
import { ServerTaskClientProvider } from '@/server-state/task';

export const App = () => (
  <ServerTaskClientProvider
    clientConfig={{
      queries: {
        staleTime: 1000 * 60 * 5, // Данные считаются свежими 5 минут
        refetchOnWindowFocus: false, // Отключаем обновление при фокусе вкладки
      },
      mutations: {
        retry: false, // Мутации при ошибке не повторяются
      },
    }}
  >
    <Router />
  </ServerTaskClientProvider>
);
```

### Чтение данных (Query)

Используйте `useServerTaskQueryAdapter`. Обратите внимание на инверсию зависимостей: вы передаете функцию `executor`, которая использует ваш сетевой модуль.

```tsx
import { useServerTaskQueryAdapter, ServerTaskQueryException } from '@/server-state/task';
import { useHttpRequester } from '@/network/http';

export const UserList = () => {
  const requester = useHttpRequester();

  const result = useServerTaskQueryAdapter({
    queryKey: ['users'],
    executor: async () => (await requester.get<User[]>('/users')).data,
  });

  // Технические флаги переведены на язык бизнеса
  if (result.isLoading) return <FullScreenLoader />; // Первичная загрузка (нет данных)
  
  if (result.isError && result.error instanceof ServerTaskQueryException) {
    // Ошибка гарантированно содержит контекст: result.error.queryKey и result.error.code
    return <ErrorFallback error={result.error} />;
  }

  return (
    <div style={{ opacity: result.isRefetching ? 0.5 : 1 }}>
      {result.data?.map(u => <div key={u.id}>{u.name}</div>)}
    </div>
  );
};
```

### Изменение данных (Mutation)

Используйте `useServerTaskMutationAdapter`. В случае ошибки в колбэке `onError` вам гарантированно доступен доменный объект ошибки с переменными, вызвавшими падение.

```tsx
import { 
  useServerTaskMutationAdapter, 
  useServerTaskClientAdapter, 
  ServerTaskMutationException 
} from '@/server-state/task';
import { useHttpRequester } from '@/network/http';

export const CreateUser = () => {
  const requester = useHttpRequester();
  const cacheActions = useServerTaskClientAdapter();

  const createMutation = useServerTaskMutationAdapter({
    mutationKey: ['create-user'],
    executor: (dto) => requester.post('/users', dto).then(r => r.data),
    
    onSuccess: (newUser) => {
      // Безопасное управление кэшем через урезанный контракт
      cacheActions.invalidateQueries(['users']);
    },
    onError: (error) => {
      if (error instanceof ServerTaskMutationException) {
        console.error(`[${error.code}] Payload с ошибкой:`, error.variables);
      }
    }
  });

  return (
    <button onClick={() => createMutation.mutate({ name: 'Иван' })} disabled={createMutation.isPending}>
      Создать
    </button>
  );
};
```

---

## ⚙️ Расширенное использование

### Понятные статусы загрузки

Модуль скрывает запутанные комбинации флагов `@tanstack/react-query` (`isPending`, `isFetching`, `fetchStatus`) и переводит их в предсказуемые бизнес-флаги:

*   **`isLoading`**: `true` **только** при самом первом запросе (когда данных в кэше еще нет). Идеально для показа скелетонов.
*   **`isRefetching`**: `true` при фоновом обновлении уже существующих данных (например, пользователь вернулся на вкладку или нажал кнопку обновления). Идеально для легкого затемнения контента, без показа спиннеров сверху.

### Безопасное управление кэшем

Хук `useServerTaskClientAdapter` возвращает контракт `IServerTaskClientActions`, который намеренно урезан. Опасные методы, нарушающие консистентность стейта (такие как `setQueryData`, позволяющий записать в кэш невалидные данные), скрыты. Доступны только безопасные операции:

*   `invalidateQueries(key)` — пометить устаревшим.
*   `removeQueries(key)` — удалить из памяти.
*   `resetQueries(key)` — сбросить в начальное состояние (очистить данные и ошибки).

### Локальное переопределение конфигурации

Хотя провайдер задает дефолтные настройки, вы можете переопределить их для конкретного запроса прямо в параметрах:

```typescript
const result = useServerTaskQueryAdapter({
  queryKey: ['heavy-report'],
  executor: () => fetchReport(),
  staleTime: 1000 * 60 * 30, // Этот кэш свежий 30 минут (вместо 5 по умолчанию)
  retry: 0, // Не повторять при ошибке
  enabled: isEnabled, // Динамическая пауза запроса
});
```

---

## 🛡️ Обработка ошибок

Модуль использует строгую иерархию исключений и выступает "Пограничным контролером" (Border Guard). Любая ошибка, выброшенная внутри функции `executor` (например, таймаут Axios, ошибка парсинга, 500 от сервера), перехватывается адаптером и **гарантированно оборачивается** в архитектурный класс-исключение.

Это позволяет извлекать контекст выполнения (`queryKey`, `variables`) без использования приведения типов (`as`) или проверок `in`.

### Иерархия исключений

Все исключения модуля наследуются от базового `IGeneralException`. Разделение на типы происходит параллельно, чтобы четко разделять контекст ошибок:

1.  **`IGeneralException` / `ServerTaskGeneralException`** — Абсолютный базовый уровень. Содержит только общие поля: `message`, `code`, `timestamp`, `details`.
2.  **`IClientException` / `ServerTaskClientException`** — Ошибка работы с клиентом кэша (например, при вызове `invalidateQueries`). Наследуется от `General`.
3.  **`IQueryException` / `ServerTaskQueryException`** — Ошибка выполнения запроса (чтение данных). Наследуется напрямую от `General`. Добавляет поле `queryKey: unknown[]`.
4.  **`IMutationException` / `ServerTaskMutationException`** — Ошибка выполнения мутации (запись данных). Наследуется напрямую от `General`. Добавляет поля `mutationKey?: unknown[]` и `variables?: unknown`.

### Пример обработки в коде:

```typescript
import { 
  ServerTaskGeneralException,
  ServerTaskQueryException,
  ServerTaskMutationException 
} from '@/server-state/task';

try {
  await someExternalAction();
} catch (error: unknown) {
  // Благодаря строгой иерархии проверки точны и безопасны
  
  if (error instanceof ServerTaskMutationException) {
    // Доступны: error.message, error.code, error.variables, error.mutationKey
    notifyUser(`Ошибка операции: ${error.message}`);
  } 
  else if (error instanceof ServerTaskQueryException) {
    // Доступны: error.message, error.code, error.queryKey
    logError(`Запрос ${error.queryKey.join('/')} упал с кодом ${error.code}`);
  }
  else if (error instanceof ServerTaskGeneralException) {
    // Базовая ошибка модуля
    logError(error.message);
  }
}
```

---

## 📚 Справочник API (Экспорты)

### Базовые типы (Contracts)
*   `IServerTaskClientConfig` — Интерфейс глобальной конфигурации (`queries.staleTime`, `mutations.retry` и т.д.). Не содержит типов TanStack Query.
*   `IServerTaskQueryParams<TData>` — Параметры запроса (`queryKey`, `executor`, `enabled`, `retry`, `staleTime`).
*   `IServerTaskQueryResult<TData>` — Результат подписки на запрос. Содержит бизнес-флаги (`data`, `isLoading`, `isRefetching`, `isError`, `error`, `refetch`).
*   `IServerTaskMutationParams<TData, TVariables>` — Параметры мутации (`mutationKey`, `executor`, `onSuccess`, `onError`).
*   `IServerTaskMutationResult<TData, TVariables>` — Результат выполнения мутации (`mutate`, `mutateAsync`, `isPending`, `data`, `isError`).
*   `IServerTaskClientActions` — Контракт императивного управления кэшем. Содержит только `invalidateQueries`, `removeQueries`, `resetQueries`.

### Типы: Контракты Исключений
*   `IGeneralException` — Базовый контракт (`message`, `code`, `timestamp`, `details?`).
*   `IClientException` — Расширяет `IGeneralException` (пустое расширение для типизации уровня клиента).
*   `IQueryException` — Расширяет `IGeneralException`. Добавляет `queryKey: unknown[]`.
*   `IMutationException` — Расширяет `IGeneralException`. Добавляет `mutationKey?` и `variables?`.

### Классы (Исключения)
Предназначены для строгих проверок через `instanceof`. Имеют правильную цепочку прототипов.
*   `ServerTaskGeneralException` — Базовая реализация.
*   `ServerTaskClientException` — Ошибки действий с кэшем в `useServerTaskClientAdapter`.
*   `ServerTaskQueryException` — Ошибки, перехваченные в `useServerTaskQueryAdapter`.
*   `ServerTaskMutationException` — Ошибки, перехваченные в `useServerTaskMutationAdapter`.

### Утилиты
*   `normalizeGeneralException<TCode>(rawError, messageFallback, codeFallback)` — Функция, извлекающая `message` и `code` из неизвестной ошибки (поддерживает стандартные `Error` и объекты с полем `code`). Используется внутри адаптеров, доступна для переиспользования.
*   `ServerTaskClientDefaultConfig` — Объект с дефолтными значениями конфигурации (staleTime: 5 мин, retry: 1, refetchOnWindowFocus: false).

### React
*   `ServerTaskClientProvider` — Провайдер контекста. Принимает `children` и `clientConfig`. Скрывает создание `QueryClientProvider`.
*   `IServerTaskClientProviderProps` — Тип пропсов провайдера.
*   `useServerTaskQueryAdapter<TData>(params)` — Хук подписки на данные. Гарантирует, что при падении `executor` выбросит `ServerTaskQueryException`.
*   `useServerTaskMutationAdapter<TData, TVariables>(params)` — Хук выполнения изменений. Гарантирует выброс `ServerTaskMutationException` при ошибке.
*   `useServerTaskClientAdapter()` — Хук для императивных действий с кэшем. Бросает ошибку при вызове вне провайдера.
*   `ServerTaskDevtools` — Экспорт компонента Devtools (обертка над `@tanstack/react-query-devtools`).
