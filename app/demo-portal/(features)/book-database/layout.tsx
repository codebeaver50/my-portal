import { Providers } from "./_components/Providers";

export default function BookDatabaseLayout({ children }: LayoutProps<"/demo-portal/book-database">) {
  return <Providers>{children}</Providers>;
}
