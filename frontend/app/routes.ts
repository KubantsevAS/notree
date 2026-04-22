import { index, route, type RouteConfig } from '@react-router/dev/routes';

export default [
  index('routes/home.tsx'),
  route(
    '/.well-known/appspecific/com.chrome.devtools.json',
    'routes/devtools-json.ts',
  ),
] satisfies RouteConfig;
