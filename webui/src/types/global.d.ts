interface Window {
  APIUrl: string
}

declare namespace JSX {
  interface IntrinsicElements {
    'hub-button-app': React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement>
    'hub-ui-demo-app': { key: string; path: string; theme: 'dark' | 'light'; baseurl: string; containercss: string }
  }
}
