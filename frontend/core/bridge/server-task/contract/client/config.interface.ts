export interface IConfig {
  queries?: {
    staleTime?: number;
    retry?: number | boolean;
    refetchOnWindowFocus?: boolean;
  };
  mutations?: {
    retry?: number | boolean;
  };
}
