import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { PageContainer } from "@/components/ui";
import { checkIsAdmin } from "@/app/_lib/adminAuth";
import { AdminLoginForm } from "./_components/AdminLoginForm";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "管理者ログイン | My Portal",
};

// ?redirect=/demo-portal/simple-cms のように、ログイン後に戻る先をfeature側から指定できる。
// 外部URLへのオープンリダイレクトを避けるため "/" 始まりの相対パスのみ許可する。
function resolveRedirectTo(value: string | string[] | undefined): string {
  if (typeof value === "string" && value.startsWith("/") && !value.startsWith("//")) {
    return value;
  }
  return "/";
}

export default async function LoginPage({ searchParams }: PageProps<"/login">) {
  const [isAdmin, params] = await Promise.all([checkIsAdmin(), searchParams]);
  const redirectTo = resolveRedirectTo(params.redirect);

  if (isAdmin) {
    redirect(redirectTo);
  }

  return (
    <PageContainer>
      <AdminLoginForm redirectTo={redirectTo} />
    </PageContainer>
  );
}
