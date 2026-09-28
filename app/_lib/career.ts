export type CareerMilestone = {
  id: string;
  period: string;
  title: string;
  description: string;
  techStack?: string[];
};

export const careerMilestones: CareerMilestone[] = [
  {
    id: "si-join",
    period: "1998年4月",
    title: "SI企業に入社",
    description:
      "新人研修を経て、Java講師（Sun認定講座・社内向け情報処理試験講座）を担当。人に教える立場から、エンジニアとしてのキャリアをスタートしました。",
  },
  {
    id: "java-lead",
    period: "1999年 – 2011年",
    title: "金融・証券系の大規模Javaシステム開発",
    description:
      "証券会社の社内システム、競馬情報提供システム、人材会社向けWebシステム、Blogシステムなど、Javaをベースにした業務システムの開発リーダー・アーキテクトとして従事。独自フレームワークの設計、大量アクセス対策、分散システム開発などを担当しました。",
    techStack: ["Java", "Spring Framework", "Oracle", "DB2", "AIX"],
  },
  {
    id: "management",
    period: "2011年 – 2018年",
    title: "ユーザー企業でのマネジメント経験",
    description:
      "スマホ向け課金システム（キャリア決済・クレジットカード決済）の開発チームをマネージャーとして統括し、後任のリーダー・アーキテクトを育成して引き継ぎを完了。2018年にかけては予防接種予約システムや子育て支援サイト・妊活アプリなど複数プロジェクトを兼任し、要件定義から運用設計までを支援しました。",
    techStack: ["Java", "AWS", "SQL Server"],
  },
  {
    id: "freelance",
    period: "2018年7月",
    title: "フリーランスとして独立",
    description:
      "「会社に依存しない働き方」を志向し独立。フルリモートで、SI企業時代の経験を生かした業務システムのフロントエンド開発を中心に活動を始めました。",
  },
  {
    id: "react-lead",
    period: "2018年 – 2021年",
    title: "React / TypeScriptでの開発リーダー経験",
    description:
      "予防接種予約システム（オフショアを含む15名体制）や子育て支援アプリの新規開発をリーダーとして担当。要件定義・アーキテクチャ設計・開発を一貫して主導しました。",
    techStack: ["C#", "TypeScript", "React", "Azure"],
  },
  {
    id: "frontend-specialist",
    period: "2021年 – 現在",
    title: "フロントエンドエンジニアとしての専業化",
    description:
      "タレントマネジメントシステムのフロントエンド開発に参画し、現在も継続中。React / TypeScriptによる設計・実装・テストを担当しています。",
    techStack: ["TypeScript", "React"],
  },
];

export const personalInterests: string[] = [
  "筋トレ: 週3回ほどジムに通っています。",
  "オンラインゲーム: FF14のライトプレイヤーです。",
  "沖縄旅行: 毎年沖縄へ旅行しています。",
  "ブラスバンド: 中学・高校で所属し、トランペットとトロンボーンを担当していました。",
  "ピアノ: 兄弟について小学生の頃に習っていました。",
];
