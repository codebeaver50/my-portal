import type { Metadata } from "next";
import { CategoryManager } from "../../_components/CategoryManager";
import { fetchCategories } from "../../_lib/api";
import type { Category } from "../../_lib/types";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "カテゴリ管理 | Simple CMS | My Portal",
};

async function loadInitialCategories(): Promise<Category[]> {
  try {
    return await fetchCategories({ cache: "no-store" });
  } catch (error) {
    // Go APIが未起動/未接続でもポータルトップ経由でページ自体は開けるようにする
    console.error("simple-cms: failed to fetch initial categories", error);
    return [];
  }
}

export default async function SimpleCmsCategoriesPage() {
  const categories = await loadInitialCategories();

  return <CategoryManager initialCategories={categories} />;
}
