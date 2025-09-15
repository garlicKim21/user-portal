/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_ARGOCD_URL: string
  readonly VITE_GRAFANA_URL: string
  // 다른 환경변수들도 여기에 추가할 수 있습니다
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
