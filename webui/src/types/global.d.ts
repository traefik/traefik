interface Window {
  APIUrl: string
}

declare namespace JSX {
  interface IntrinsicElements {
    'hub-button-app': React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement>
  }
}
