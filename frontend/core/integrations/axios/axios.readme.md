# Axios client registry

Модуль `core/injections/axios` дает единый способ регистрировать и получать `axios`-клиенты через React Context.

## Что экспортируется

Из `../../core/injections`:

- `AxiosClientRegistry` — класс-реестр клиентов.
- `AxiosClientProvider` — provider, который кладет `AxiosClientRegistry` в контекст.
- `useAxiosClient(clientKey?)` — хук для получения клиента по ключу.

## Как это устроено

- `AxiosClientRegistry` хранит:
  - `defaultClientKey` (по умолчанию `'default'`);
  - внутренний `storage` с клиентами.
- `AxiosClientProvider` принимает один из вариантов:
  - `registry` (рекомендуется),
  - `clients`,
  - `client`.
- `useAxiosClient()`:
  - берет клиент по ключу;
  - если ключ не передан — использует `defaultClientKey` из реестра;
  - если клиент не найден — кидает ошибку с доступными ключами.

## Текущая интеграция в проекте

### 1) Создание реестра и default клиента

`frontend/app/injections/axios.injection.tsx`

```tsx
import axios from 'axios';
import { AxiosClientProvider, AxiosClientRegistry } from '../../core/injections';

const defaultClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
});

const axiosClientRegistry = new AxiosClientRegistry();
axiosClientRegistry.setDefaultClientKey('default');
axiosClientRegistry.createStorage();
axiosClientRegistry.registerClient(defaultClient);
```

### 2) Подключение provider

```tsx
export function AxiosInjection({ children }: { children: React.ReactNode }) {
  return (
    <AxiosClientProvider registry={axiosClientRegistry}>
      {children}
    </AxiosClientProvider>
  );
}
```

### 3) Использование клиента

```ts
import { useAxiosClient } from '../../core/injections';

export function useUsersApi() {
  const client = useAxiosClient(); // default client

  return {
    getUsers: () => client.get('/users'),
  };
}
```

## Несколько клиентов

```ts
const authClient = axios.create({
  baseURL: import.meta.env.VITE_AUTH_API_URL,
});

axiosClientRegistry.registerClient(authClient, 'auth');
```

И получение по ключу:

```ts
const authClient = useAxiosClient('auth');
```

## Методы `AxiosClientRegistry`

- `setDefaultClientKey(key)`
- `getDefaultClientKey()`
- `createStorage()`
- `setStorage(registry)`
- `getStorage()`
- `registerClient(client, key?)`
- `registerClients(clients)`
- `getClient(key)`
- `hasClient(key)`
- `getRegisteredKeys()`

## Рекомендации

- Создавайте один `AxiosClientRegistry` на уровне приложения.
- Явно задавайте default key при инициализации.
- Используйте понятные ключи (`'default'`, `'auth'`, `'billing'`) для доменных API.
