import type { Metadata } from "next";
import { Grid, PageContainer, PageHeader } from "@/components/ui";
import { CaseStudyCard } from "../_components/CaseStudyCard";
import { caseStudies } from "../_lib/works";

export const metadata: Metadata = {
  title: "Works | Code Beaver",
  description: "業務委託として携わった実務案件の事例です。",
};

export default function WorksPage() {
  return (
    <PageContainer>
      <PageHeader
        title="Works"
        description="業務委託として携わった案件の一部です（守秘義務の範囲で匿名化しています）。"
      />
      <Grid columns={2}>
        {caseStudies.map((caseStudy) => (
          <CaseStudyCard key={caseStudy.id} caseStudy={caseStudy} />
        ))}
      </Grid>
    </PageContainer>
  );
}
