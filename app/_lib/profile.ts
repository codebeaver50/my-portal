export type Reason = {
  title: string;
  description: string;
};

export type SupportItem = {
  title: string;
  description: string;
};

export const heroCopy = {
  title: ["業務理解を大切にする", "React / TypeScriptエンジニア"],
  description:
    "SI企業でのバックエンド開発（約20年）を経て、2018年からはフリーランスとしてReact / TypeScriptを中心としたフロントエンド開発に従事しています。開発チームの一員として、業務システムの機能追加・改善から既存画面のReact移行まで対応します。",
};

export const reasons: Reason[] = [
  {
    title: "業務理解",
    description: "SI企業での約20年の経験を生かし、複雑な業務要件を整理して設計に落とし込みます。",
  },
  {
    title: "システム全体の理解",
    description: "API・DB・バックエンドまで踏まえたうえで、フロントエンドの設計・実装を行います。",
  },
  {
    title: "継続的な開発支援",
    description: "常駐先でのチーム開発や、既存システムの機能追加・保守改善に長期で対応します。",
  },
];

export const supportItems: SupportItem[] = [
  {
    title: "業務システムのフロントエンド開発",
    description: "業務要件を理解したうえで、React / TypeScriptによる画面の設計・実装を担当します。",
  },
  {
    title: "既存システムの機能追加・改善",
    description: "既存コードを読み解き、機能追加・UI改善・リファクタリングを継続的に行います。",
  },
  {
    title: "レガシー画面のReact移行",
    description: "JSPなどで作られた既存の画面を、React / TypeScriptへ段階的に移行します。",
  },
  {
    title: "API連携・仕様調整",
    description: "バックエンドの経験を生かし、API仕様の整理からフロントエンドとの連携まで、バックエンド担当者と調整しながら進めます。",
  },
];

export const consultationExamples: string[] = [
  "React / TypeScriptのエンジニアを開発チームに加えたい",
  "既存のReactシステムに継続的に機能を追加してほしい",
  "複雑になったフロントエンドのコードを整理したい",
  "JSPなどで作られた古い画面をReactへ置き換えたい",
  "仕様・デザインのレビューから関わってほしい",
  "バックエンドも理解しているフロントエンドエンジニアを探している",
];

export const currentFocus: string[] = [
  "React",
  "TypeScript",
  "Next.js",
  "フロントエンド設計",
  "状態管理・API連携",
  "テスト・保守改善",
];

export const foundationSkills: string[] = [
  "Java / JSP",
  "Tomcat",
  "MySQL / Oracle / SQL Server",
  "API・DB設計",
  "Webシステム全体の理解",
];

export const supportedWork: string[] = [
  "新規開発",
  "既存システム改善",
  "レガシー刷新",
  "バックエンド担当者との仕様調整",
];

export const careerSummary =
  "SI企業で約20年、Javaを中心としたバックエンド開発に従事。2018年にフリーランスとして独立し、以降はReact / TypeScriptを中心としたフロントエンド開発に約8年携わっています。";

export const contactInfo = {
  x: "https://x.com/codebeaver50",
  github: "https://github.com/codebeaver50",
};
