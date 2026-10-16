import type { Metadata } from "next";
import { PageContainer } from "@/components/ui";
import { resolveRedirectTo } from "../_lib/auth";
import { LoginForm } from "./_components/LoginForm";

export const metadata: Metadata = {
  title: "ログイン | DemoPortal | Code Beaver",
};

// ?redirect=/demo-portal/simple-ledger のように、ログイン後に戻る先をfeature側から指定できる。
export default async function DemoPortalLoginPage({ searchParams }: PageProps<"/demo-portal/login">) {
  const params = await searchParams;

  return (
    <PageContainer>
      <LoginForm redirectTo={resolveRedirectTo(params.redirect)} />
    </PageContainer>
  );
}
