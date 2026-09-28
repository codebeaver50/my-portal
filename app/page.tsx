import { Card, CardDescription, CardHeader, CardTitle, Grid, LinkButton, PageContainer, TechTag } from "@/components/ui";
import { CaseStudyCard } from "./_components/CaseStudyCard";
import {
  careerSummary,
  consultationExamples,
  currentFocus,
  foundationSkills,
  heroCopy,
  reasons,
  supportItems,
} from "./_lib/profile";
import { caseStudies } from "./_lib/works";

const HIGHLIGHTED_CASE_STUDY_IDS = ["talent-management-frontend", "seminar-cms"];

export default function Home() {
  const highlightedCaseStudies = HIGHLIGHTED_CASE_STUDY_IDS.map((id) =>
    caseStudies.find((caseStudy) => caseStudy.id === id),
  ).filter((caseStudy) => caseStudy !== undefined);

  return (
    <PageContainer>
      <section className="flex flex-col gap-6 pb-8">
        <h1 className="text-2xl leading-tight font-extrabold tracking-tight sm:text-3xl">
          {heroCopy.title.map((line) => (
            <span key={line} className="block">
              {line}
            </span>
          ))}
        </h1>
        <p className="max-w-3xl text-base text-muted-foreground">{heroCopy.description}</p>
        <div className="flex flex-col gap-3 sm:flex-row">
          <LinkButton href="/works" variant="accent" className="w-auto">
            開発実績を見る →
          </LinkButton>
          <LinkButton href="/contact" variant="secondary" className="w-auto">
            仕事について相談する
          </LinkButton>
        </div>
      </section>

      <section className="flex flex-col gap-6 pb-8">
        <h2 className="text-2xl font-extrabold tracking-tight">対応できること</h2>
        <Grid columns={2}>
          {supportItems.map((item) => (
            <Card key={item.title}>
              <CardHeader>
                <CardTitle>{item.title}</CardTitle>
                <CardDescription>{item.description}</CardDescription>
              </CardHeader>
            </Card>
          ))}
        </Grid>
        <h3 className="text-lg font-bold">こんなご相談に対応します</h3>
        <ul className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          {consultationExamples.map((example) => (
            <li key={example} className="flex items-start gap-2 text-sm text-muted-foreground">
              <span className="mt-1.5 size-1.5 flex-none rounded-full bg-accent" aria-hidden />
              {example}
            </li>
          ))}
        </ul>
      </section>

      <section className="flex flex-col gap-6 pb-8">
        <h2 className="text-2xl font-extrabold tracking-tight">選ばれる理由</h2>
        <Grid columns={3}>
          {reasons.map((reason) => (
            <Card key={reason.title}>
              <CardHeader>
                <CardTitle>{reason.title}</CardTitle>
                <CardDescription>{reason.description}</CardDescription>
              </CardHeader>
            </Card>
          ))}
        </Grid>
      </section>

      <section className="flex flex-col gap-6 pb-8">
        <h2 className="text-2xl font-extrabold tracking-tight">直近の実績</h2>
        <Grid columns={2}>
          {highlightedCaseStudies.map((caseStudy) => (
            <CaseStudyCard key={caseStudy.id} caseStudy={caseStudy} />
          ))}
        </Grid>
        <LinkButton href="/works" variant="secondary" className="w-auto">
          その他のWorksを見る
        </LinkButton>
      </section>

      <section className="flex flex-col gap-6 pb-8">
        <h2 className="text-2xl font-extrabold tracking-tight">スキル</h2>
        <Grid columns={2}>
          <Card>
            <CardHeader>
              <CardTitle>現在の専門領域</CardTitle>
            </CardHeader>
            <div className="flex flex-wrap gap-1.5">
              {currentFocus.map((item) => (
                <TechTag key={item}>{item}</TechTag>
              ))}
            </div>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>技術的基盤</CardTitle>
            </CardHeader>
            <div className="flex flex-wrap gap-1.5">
              {foundationSkills.map((item) => (
                <TechTag key={item}>{item}</TechTag>
              ))}
            </div>
          </Card>
        </Grid>
        <LinkButton href="/skills" variant="secondary" className="w-auto">
          スキルの詳細を見る
        </LinkButton>
      </section>

      <section className="flex flex-col gap-6 pb-8">
        <h2 className="text-2xl font-extrabold tracking-tight">経歴</h2>
        <p className="max-w-3xl text-base text-muted-foreground">{careerSummary}</p>
        <LinkButton href="/profile" variant="secondary" className="w-auto">
          経歴の詳細を見る
        </LinkButton>
      </section>

      <section className="flex flex-col gap-4 rounded-xl border border-border bg-card p-8 text-center">
        <h2 className="text-2xl font-extrabold tracking-tight">お仕事のご相談はこちらから</h2>
        <div className="flex justify-center">
          <LinkButton href="/contact" variant="accent" className="w-auto">
            お問い合わせ →
          </LinkButton>
        </div>
      </section>
    </PageContainer>
  );
}
