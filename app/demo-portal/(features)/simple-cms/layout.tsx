import { PageContainer } from "@/components/ui";
import { Providers } from "./_components/Providers";

export default function SimpleCmsLayout({ children }: LayoutProps<"/demo-portal/simple-cms">) {
  return (
    <Providers>
      <PageContainer>{children}</PageContainer>
    </Providers>
  );
}
