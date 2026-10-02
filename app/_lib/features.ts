export type Feature = {
  slug: string;
  name: string;
  description: string;
  category: string;
  status: "available" | "coming-soon";
  techStack: string[];
};

export const features: Feature[] = [
  {
    slug: "book-database",
    name: "書籍データベース",
    description: "国立国会図書館サーチ・openBDから収集した今月発売されるIT関連書籍の一覧です。",
    category: "DATABASE SYSTEM",
    status: "available",
    techStack: ["Next.js", "Go", "MySQL", "NDL Search API", "openBD API"],
  },
  {
    slug: "simple-ledger",
    name: "複式簿記シンプル家計簿",
    description: "複式簿記ベースの取引管理・分析ができるシンプルな家計簿アプリです。",
    category: "LEDGER SYSTEM",
    status: "available",
    techStack: ["Next.js", "Go", "MySQL"],
  },
  {
    slug: "simple-cms",
    name: "Simple CMS",
    description: "カテゴリ分けとリッチテキスト編集ができる、シンプルな記事投稿・閲覧アプリです。",
    category: "CMS",
    status: "available",
    techStack: ["Next.js", "Go", "MySQL", "Tiptap"],
  },
  {
    slug: "form-builder",
    name: "カスタムフォームビルダー",
    description: "項目を自由に組み合わせてフォームを作成し、集まった回答をデータベースとして一覧できるアプリです。",
    category: "FORM & DATABASE",
    status: "available",
    techStack: ["Next.js", "Go", "MySQL"],
  },
  {
    slug: "shop-css",
    name: "Antiqueshop Oude dagen",
    description: "アンティークショップを想定した、HTML/CSSで制作した静的サイトのポートフォリオです。",
    category: "STATIC SITE",
    status: "available",
    techStack: ["HTML", "CSS"],
  },
];
