import type { Metadata } from "next";
import { AnswerForm } from "../_components/AnswerForm";
import { FormHeader } from "../_components/FormHeader";
import { getForm } from "../_lib/getForm";

export const dynamic = "force-dynamic";

export async function generateMetadata({ params }: PageProps<"/form-builder/[formId]">): Promise<Metadata> {
  const { formId } = await params;
  const form = await getForm(formId);
  return { title: `${form.title} | カスタムフォームビルダー | My Portal` };
}

export default async function AnswerFormPage({ params }: PageProps<"/form-builder/[formId]">) {
  const { formId } = await params;
  const form = await getForm(formId);

  return (
    <div className="flex flex-col gap-6">
      <FormHeader form={form} />
      <AnswerForm form={form} />
    </div>
  );
}
