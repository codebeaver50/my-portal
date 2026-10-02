import type { Metadata } from "next";
import { FormHeader } from "../../_components/FormHeader";
import { RecordsTable } from "../../_components/RecordsTable";
import { getForm } from "../../_lib/getForm";

export const dynamic = "force-dynamic";

export async function generateMetadata({ params }: PageProps<"/demo-portal/form-builder/[formId]/records">): Promise<Metadata> {
  const { formId } = await params;
  const form = await getForm(formId);
  return { title: `${form.title}の回答データ | カスタムフォームビルダー | My Portal` };
}

export default async function FormRecordsPage({ params }: PageProps<"/demo-portal/form-builder/[formId]/records">) {
  const { formId } = await params;
  const form = await getForm(formId);

  return (
    <div className="flex flex-col gap-6">
      <FormHeader form={form} />
      <RecordsTable form={form} />
    </div>
  );
}
