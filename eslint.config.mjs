import { readdirSync } from "node:fs";
import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";

// fe_coding_rules.md のうち、機械的にチェックできるルールをここで強制する。
// ルールを追加・変更した場合は fe_coding_rules.md も揃えること。

const featuresDir = "./app/demo-portal/(features)";
const featureSlugs = readdirSync(featuresDir, { withFileTypes: true })
  .filter((entry) => entry.isDirectory())
  .map((entry) => entry.name);

// Next.js が default export を要求するファイル
const nextSpecialFiles = [
  "page",
  "layout",
  "template",
  "loading",
  "error",
  "global-error",
  "not-found",
  "default",
  "route",
].map((name) => `app/**/${name}.{ts,tsx}`);

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  {
    linterOptions: {
      // 不要になった eslint-disable を残さない
      reportUnusedDisableDirectives: "error",
    },
    rules: {
      // useEffect の依存配列をごまかさない（止める場合は理由をコメントに書いて eslint-disable する）
      "react-hooks/exhaustive-deps": "error",
      // 型定義は interface ではなく type
      "@typescript-eslint/consistent-type-definitions": ["error", "type"],
      // enum ではなく文字列リテラルのユニオン型
      "no-restricted-syntax": [
        "error",
        {
          selector: "TSEnumDeclaration",
          message: "enum ではなく文字列リテラルのユニオン型を使ってください",
        },
      ],
      // 型だけの読み込みは import type（`import { type X }` ではなく行を分ける）
      "@typescript-eslint/consistent-type-imports": [
        "error",
        { prefer: "type-imports", fixStyle: "separate-type-imports" },
      ],
      "import/consistent-type-specifier-style": ["error", "prefer-top-level"],
      // feature から別の feature を import しない
      "import/no-restricted-paths": [
        "error",
        {
          zones: featureSlugs.map((slug) => ({
            target: `${featuresDir}/${slug}`,
            from: featuresDir,
            except: [`./${slug}`],
            message:
              "feature 間の依存は禁止です。共通の処理は components/ui か app/demo-portal/_lib に移してください",
          })),
        },
      ],
    },
  },
  {
    // export は named export にする（Next.js が要求するファイルは除く）
    files: ["app/**/*.{ts,tsx}", "components/**/*.{ts,tsx}"],
    ignores: nextSpecialFiles,
    rules: {
      "import/no-default-export": "error",
    },
  },
  // Override default ignores of eslint-config-next.
  globalIgnores([
    // Default ignores of eslint-config-next:
    ".next/**",
    "out/**",
    "build/**",
    "next-env.d.ts",
  ]),
]);

export default eslintConfig;
