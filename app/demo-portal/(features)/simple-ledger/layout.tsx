import { PageContainer, PageHeader } from "@/components/ui";
import { FeatureNav } from "./_components/FeatureNav";
import { Providers } from "./_components/Providers";

export default function SimpleLedgerLayout({ children }: LayoutProps<"/demo-portal/simple-ledger">) {
  return (
    <Providers>
      <PageContainer>
        <PageHeader
          title="複式簿記シンプル家計簿"
          description="複式簿記の仕組みで取引を記録・分析するデモアプリです。画面の操作は誰でも試せますが、保存・削除には編集権限のあるアカウントでのログインが必要です。"
        />
        <FeatureNav />
        {children}
      </PageContainer>
    </Providers>
  );
}
