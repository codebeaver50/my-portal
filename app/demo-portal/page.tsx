import type { Metadata } from "next";
import { Grid, PageContainer, PageHeader } from "@/components/ui";
import { FeatureCard } from "./_components/FeatureCard";
import { PortalStats } from "./_components/PortalStats";
import { features } from "../_lib/features";

export const metadata: Metadata = {
  title: "DemoPortal | Code Beaver",
  description: "このポートフォリオサイト上で実際に起動・閲覧できるデモアプリケーションの一覧です。",
};

export default function PortalPage() {
  return (
    <PageContainer>
      <PageHeader
        title="DemoPortal"
        description="このポートフォリオサイト上で実際に起動・閲覧できるデモアプリケーション（feature）の一覧です。"
        aside={<PortalStats features={features} />}
      />
      <Grid columns={2}>
        {features.map((feature) => (
          <FeatureCard key={feature.slug} feature={feature} />
        ))}
      </Grid>
    </PageContainer>
  );
}
