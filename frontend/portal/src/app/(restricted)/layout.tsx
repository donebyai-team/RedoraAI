import { PortalAuthProvider } from '../../provider/PortalAuthProvider'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (<PortalAuthProvider>{children}</PortalAuthProvider>);
}
