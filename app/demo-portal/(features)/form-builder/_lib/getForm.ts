import { cache } from "react";
import { notFound } from "next/navigation";
import { fetchForm } from "./api";
import type { Form } from "./types";

// getForm は Server Component 用に、URLの formId からフォームを取得する。
// generateMetadata とページ本体で同じフォームを取得するため React の cache で
// 1リクエスト内の重複取得をまとめる。不正なID・存在しないフォームは404にする。
export const getForm = cache(async (formIdParam: string): Promise<Form> => {
  const formId = Number(formIdParam);
  if (!Number.isInteger(formId) || formId < 1) {
    notFound();
  }

  const form = await fetchForm(formId, { cache: "no-store" });
  if (!form) {
    notFound();
  }
  return form;
});
