import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 30_000,
    },
  },
});

export function ReactQueryInjection({
  children,
}: {
  children: React.ReactNode;
}) {
  const isDevMode = import.meta.env.MODE === 'development';

  return (
    <QueryClientProvider client={queryClient}>
      {isDevMode && <ReactQueryDevtools initialIsOpen={false} />}
      {children}
    </QueryClientProvider>
  );
}
