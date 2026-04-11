import type { Route } from './+types/home';
import { Welcome } from '../../pages/welcome';

// eslint-disable-next-line no-empty-pattern
export function meta({}: Route.MetaArgs) {
  return [{ title: 'New React Router App' }, { name: 'description', content: 'Welcome to React Router!' }];
}

export default function Home() {
  return <Welcome />;
}
