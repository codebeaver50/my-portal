import type { Metadata } from "next";
import { FormBuilder } from "../../_components/builder/FormBuilder";
import { FormHeader } from "../../_components/FormHeader";
import { getForm } from "../../_lib/getForm";

export const dynamic = "force-dynamic";

export async function generateMetadata({ params }: PageProps<"/form-builder/[formId]/edit">): Promise<Metadata> {
  const { formId } = await params;
  const form = await getForm(formId);
  return { title: `${form.title}を編集 | カスタムフォームビルダー | My Portal` };
}

export default async function EditFormPage({ params }: PageProps<"/form-builder/[formId]/edit">) {
  const { formId } = await params;
  const form = await getForm(formId);

  return (
    <div className="flex flex-col gap-6">
      <FormHeader form={form} />
      <FormBuilder key={form.updatedAt} form={form} />
    </div>
  );
}
