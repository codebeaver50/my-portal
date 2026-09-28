import type { Metadata } from "next";
import { Card, CardDescription, CardHeader, CardTitle, PageContainer, PageHeader, ToastProvider } from "@/components/ui";
import { ContactForm } from "./_components/ContactForm";

export const metadata: Metadata = {
  title: "Contact | Code Beaver",
  description: "お仕事のご相談・お問い合わせはこちらから。",
};

export default function ContactPage() {
  return (
    <PageContainer>
      <PageHeader title="Contact" description="お仕事のご相談はお気軽にご連絡ください。" />

      <Card>
        <CardHeader>
          <CardTitle>対応可能な条件</CardTitle>
        </CardHeader>
        <ul className="flex flex-col gap-2 text-sm text-muted-foreground">
          <li>契約形態: 業務委託（準委任）</li>
          <li>稼働条件: フルタイム・一部稼働ともに応相談</li>
          <li>リモート対応: フルリモート対応可（実績多数）</li>
        </ul>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>こんな段階でもご相談ください</CardTitle>
          <CardDescription>正式な打診の前でも構いません。まずはお気軽にご連絡ください。</CardDescription>
        </CardHeader>
        <ul className="flex flex-col gap-2 text-sm text-muted-foreground">
          <li>案件の詳細や開発体制がまだ固まっていない</li>
          <li>参画時期や稼働量が未定</li>
          <li>スキルが案件に合うか確認したい</li>
          <li>まずは技術的な相談をしたい</li>
        </ul>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>問い合わせフォーム</CardTitle>
          <CardDescription>下記フォームからのご連絡はメールで受け取ります。</CardDescription>
        </CardHeader>
        <ToastProvider>
          <ContactForm />
        </ToastProvider>
      </Card>
    </PageContainer>
  );
}
