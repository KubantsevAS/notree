import '../core/ui/app.css';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { useEffect } from 'react';
import { Links, Meta, Outlet, Scripts, ScrollRestoration } from 'react-router';

import { HttpAxiosProvider, useHttpAxiosRequester } from '../core/network';
import type { Route } from './+types/root';

export const links: Route.LinksFunction = () => [
  { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
  {
    rel: 'preconnect',
    href: 'https://fonts.gstatic.com',
    crossOrigin: 'anonymous',
  },
  {
    rel: 'stylesheet',
    href: 'https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap',
  },
];

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang='en'>
      <head>
        <meta charSet='utf-8' />
        <meta name='viewport' content='width=device-width, initial-scale=1' />
        <Meta />
        <Links />
      </head>
      <body>
        {children}
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}

function HttpAxiosRequesterTest() {
  const requester = useHttpAxiosRequester();

  useEffect(() => {
    const todo = async () => {
      const result = await requester.get('/todos/1');
      console.log(result);
      return result;
    };

    todo();
  });

  return 'HttpRequesterTest';
}

export default function App() {
  const queryClient = new QueryClient();

  const isDevMode = import.meta.env.MODE === 'development';

  return (
    <HttpAxiosProvider
      config={{ baseURL: 'https://jsonplaceholder.typicode.com' }}
    >
      <QueryClientProvider client={queryClient}>
        {isDevMode && <ReactQueryDevtools initialIsOpen={false} />}
        <HttpAxiosRequesterTest />
        <Outlet />
      </QueryClientProvider>
    </HttpAxiosProvider>
  );
}
