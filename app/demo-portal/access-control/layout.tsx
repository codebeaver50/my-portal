import { Providers } from "./_components/Providers";

export default function AccessControlLayout({ children }: LayoutProps<"/demo-portal/access-control">) {
  return <Providers>{children}</Providers>;
}
