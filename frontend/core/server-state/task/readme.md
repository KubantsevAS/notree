# Документация модуля `server-state/task`

Модуль предназначен для управления серверным состоянием (кэшированием, фоновым обновлением, синхронизацией) в React-приложениях. 

Он полностью изолирует бизнес-логику приложения от конкретной библиотеки-реализации (в текущей реализации используется TanStack Query). Прямой импорт из `@tanstack/react-query` в компонентах строго запрещен.

---

## 🏗 Архитектурные принципы

1. **Контракты вместо реализации:** Компоненты работают с интерфейсами `ITaskQueryParams` и `ITaskQueryResult`, описывающими *что* нужно сделать, а не *как*.
2. **Инверсия зависимостей (executor):** Модуль не знает, как отправлять HTTP-запросы. Сетевой модуль (`network/http`) пробрасывается внутрь через функцию `executor` в момент вызова хука.
3. **Безопасное управление кэшем:** Императивное управление кэшем (`useTaskClientAdapter`) доступно только через урезанный контракт `ITaskClientActions`. Опасные методы (такие как `setQueryData`, часто приводящие к рассинхронизации) намеренно скрыты.
4. **Понятные статусы:** Технические флаги состояния TanStack Query переведены на язык бизнеса: `isLoading` (первичная загрузка) отделено от `isRefetching` (фоновое обновление).

---

## 🚀 Быстрый старт

### 1. Инициализация (Провайдер)

Оберните корень приложения в `TaskClientProvider`. Он инициализирует внутренний клиент с глобальными дефолтами.

```tsx
import { TaskClientProvider } from '@/server-state/task';

export const App = () => (
  <TaskClientProvider
    clientConfig={{
      queries: {
        staleTime: 1000 * 60 * 5, // Данные свежие 5 минут
        refetchOnWindowFocus: false, // Не обновлять при возврате на вкладку
      },
      mutations: {
        retry: false, // Не повторять POST/PUT запросы при ошибке
      },
    }}
  >
    <Router />
  </TaskClientProvider>
);
```

### 2. Чтение данных (Query)

Используйте `useTaskQueryAdapter`. Обратите внимание на связку с сетевым модулем через `executor`.

```tsx
import { useTaskQueryAdapter, type ITaskQueryResult } from '@/server-state/task';
import { useHttpRequester } from '@/network';

export const UserList = () => {
  const requester = useHttpRequester();

  const result: ITaskQueryResult<User[]> = useTaskQueryAdapter({
    queryKey: ['users'],
    // Связываем кэш и сеть
    executor: async () => (await requester.get<User[]>('/users')).data,
  });

  if (result.isLoading) return <FullScreenLoader />; // Нет данных вообще
  if (result.isError) return <ErrorFallback error={result.error} />;

  return (
    <div style={{ opacity: result.isRefetching ? 0.5 : 1 }}>
      {/* isRefetching = true, если данные есть, но идет фоновое обновление */}
      {result.data?.map(u => <div key={u.id}>{u.name}</div>)}
    </div>
  );
};
```

### 3. Изменение данных (Mutation)

Используйте `useTaskMutationAdapter` для создания/обновления сущностей.

```tsx
import { useTaskMutationAdapter, useTaskClientAdapter } from '@/server-state/task';

export const CreateUser = () => {
  const requester = useHttpRequester();
  const cacheActions = useTaskClientAdapter(); // Для сброса кэша

  const createMutation = useTaskMutationAdapter({
    mutationKey: ['create-user'],
    executor: (dto) => requester.post('/users', dto).then(r => r.data),
    
    onSuccess: (newUser) => {
      // Безопасно инвалидируем список пользователей после успеха
      cacheActions.invalidateQueries(['users']);
    },
    onError: (error) => {
      console.error('Ошибка создания:', error);
    }
  });

  const handleCreate = () => {
    createMutation.mutate({ name: 'Иван' });
  };

  return (
    <button onClick={handleCreate} disabled={createMutation.isPending}>
      {createMutation.isPending ? 'Создание...' : 'Создать пользователя'}
    </button>
  );
};
```

---

## 📚 Справочник API

### Хуки (Адаптеры)

| Экспорт | Описание |
| :--- | :--- |
| `useTaskQueryAdapter<TData, TError>(params)` | Хук для подписки на данные. Возвращает `ITaskQueryResult`. |
| `useTaskMutationAdapter<TData, TVariables, TError>(params)` | Хук для выполнения изменений. Возвращает `ITaskMutationResult`. |
| `useTaskClientAdapter()` | Хук для императивного управления кэшем. Возвращает `ITaskClientActions`. |

### Провайдеры

| Экспорт | Описание |
| :--- | :--- |
| `TaskClientProvider` | Компонент-обертка. Принимает пропсы `children` и `clientConfig` типа `ITaskClientConfig`. |

### Типы: Контракты Query

*   `ITaskQueryParams<TData, TError>`
    *   `queryKey: unknown[]` — Ключ кэша.
    *   `executor: () => Promise<TData>` — Функция запроса данных.
    *   `enabled?: boolean` — Управление паузой запроса.
    *   `retry?: number` — Локальное переопределение количества попыток.
    *   `staleTime?: number` — Локальное переопределение времени "свежести" данных.
*   `ITaskQueryResult<TData, TError>`
    *   `data: TData | undefined` — Данные из кэша.
    *   `isLoading: boolean` — `true` только при самом первом запросе (когда `data === undefined`).
    *   `isRefetching: boolean` — `true` при фоновом обновлении уже существующих данных.
    *   `isError: boolean` / `error: TError | null` — Флаг и объект ошибки.
    *   `refetch: () => void` — Императивный вызов рефетча.

### Типы: Контракты Mutation

*   `ITaskMutationParams<TData, TVariables, TError>`
    *   `mutationKey?: unknown[]` — Ключ для дебаунса одинаковых мутаций.
    *   `executor: (vars: TVariables) => Promise<TData>` — Функция отправки данных.
    *   `onSuccess?: (data, vars) => void` / `onError?: (error, vars) => void` — Коллбэки life-cycle.
*   `ITaskMutationResult<TData, TError, TVariables>`
    *   `mutate: (vars) => void` — Триггер мутации (fire-and-forget).
    *   `mutateAsync: (vars) => Promise<TData>` — Триггер с ожиданием промиса.
    *   `isPending: boolean` — Мутация выполняется в данный момент.
    *   `data`, `isError`, `error` — Результат выполнения.

### Типы: Управление клиентом

*   `ITaskClientConfig` — Глобальные настройки (передаются в Провайдер).
*   `ITaskClientActions`
    *   `invalidateQueries(key)` — Помечает кэш по ключу как устаревший (вызывает рефетч активных подписок).
    *   `removeQueries(key)` — Полностью удаляет данные из кэша.
    *   `resetQueries(key)` — Удаляет данные и сбрасывает статусы (isError и т.д.) в начальное состояние.
