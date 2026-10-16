# Frontend Coding Rules（React / TypeScript / Next.js）

このドキュメントには、一般的なスタイルガイドに加えて **このプロジェクトで決めたこと** と **避けたいアンチパターン** を書く。
ディレクトリ構成・設計判断は `ARCHITECTURE.md` を正とし、ここではコードの書き方を扱う。

## 前提とするガイド

ここに書いていないことは、以下に従う。

- [Rules of React](https://react.dev/reference/rules)（コンポーネント・Hooksの純粋性、Hooksのルール）
- [You Might Not Need an Effect](https://react.dev/learn/you-might-not-need-an-effect)（`useEffect` の誤用パターン集）
- [Google TypeScript Style Guide](https://google.github.io/styleguide/tsguide.html)
- Next.js 16 の公式ドキュメント（`node_modules/next/dist/docs/`）。学習データの知識と異なる部分があるため、APIを使う前に確認する

機械的にチェックできるものは ESLint（`npm run lint`）・TypeScript の `strict` に任せる。CI（`.github/workflows/ci.yml`）でも `npm run lint` を実行し、error のルールに違反するとCIが失敗する。

### Lint で強制しているルール

`eslint.config.mjs` で、このドキュメントのうち以下のルールを error にしている。ルールを追加・変更した場合は、このドキュメントと `eslint.config.mjs` の両方を揃える。

| ルール | ESLint のルール |
|---|---|
| `interface` ではなく `type` を使う | `@typescript-eslint/consistent-type-definitions` |
| `enum` を使わない | `no-restricted-syntax`（`TSEnumDeclaration`） |
| `any` を使わない | `@typescript-eslint/no-explicit-any` |
| 型だけの読み込みは `import type`（行を分ける） | `@typescript-eslint/consistent-type-imports` / `import/consistent-type-specifier-style` |
| default export は Next.js が要求するファイルだけ | `import/no-default-export`（`page` / `layout` 等を除く） |
| feature から別の feature を import しない | `import/no-restricted-paths` |
| useEffect の依存配列をごまかさない | `react-hooks/exhaustive-deps` |
| useEffect の中で同期的に `setState` しない | `react-hooks/set-state-in-effect` |
| 不要になった `eslint-disable` を残さない | `linterOptions.reportUnusedDisableDirectives` |

`eslint-disable` に理由のコメントがあるかは Lint では確認していないため、コードレビューで確認する。

## TypeScript

### 型

- 型定義は `type` を使う（`interface` は使わない）。取りうる値の列挙は `enum` ではなく文字列リテラルのユニオン型にする
  ```ts
  // ✅
  export type EntryType = "debit" | "credit";
  // ❌
  export enum EntryType { Debit = "debit", Credit = "credit" }
  ```
- `any` を使わない。型が分からない値は `unknown` で受けて、絞り込んでから使う
  ```ts
  // ✅
  const onError = (error: unknown) => {
    showToast(error instanceof Error ? error.message : "保存に失敗しました", "error");
  };
  // ❌
  const onError = (error: any) => showToast(error.message, "error");
  ```
- `as` による型アサーション・非null アサーション（`!`）は原則使わない。使う場合は、なぜ安全なのかをコメントに書く
- 値ごとの対応表は `Record<ユニオン型, 値>` にして、値の追加時に対応漏れを型エラーで検出できるようにする
  ```ts
  const ACCOUNT_TYPE_LABEL: Record<ChartOfAccount["type"], string> = { asset: "資産", /* ... */ };
  ```
- 「取得中」「失敗」「成功」のように状態ごとに持つデータが変わる場合は、フラグを複数持たずに判別可能なユニオン型にする
  ```ts
  // ✅
  type LoadState =
    | { status: "loading" }
    | { status: "error"; message: string }
    | { status: "loaded"; data: RolesResponse };
  // ❌ isLoading と error と data の組み合わせで、ありえない状態が表現できてしまう
  ```

### import / export

- 型だけを読み込むときは `import type` を使う
- export は named export にする。`default export` は Next.js が要求するファイル（`page.tsx` / `layout.tsx` 等）だけ
- 別のディレクトリ（`components/ui`・`app/demo-portal/_lib` 等）は `@/` のエイリアスで、同じfeatureの中は相対パスで読み込む

## React

### コンポーネント

- コンポーネントは `function` 宣言で書く（`React.FC` やアロー関数の代入にしない）。propsの型は `<コンポーネント名>Props` として同じファイルに定義する
- 1ファイル1コンポーネントを基本とする。そのコンポーネントでしか使わない小さな補助コンポーネント・関数は同じファイルに置いてよい
- レンダリング中に計算できる値を state に持たない（後述の `useEffect` のアンチパターンも参照）
  ```ts
  // ✅ lines から毎回計算する
  const debitTotal = lines.filter((line) => line.type === "debit").reduce(/* ... */);
  // ❌ lines が変わるたびに useEffect で setDebitTotal する
  ```
- 計算が重い・参照の同一性が必要な場合だけ `useMemo` / `useCallback` を使う。とりあえず付けることはしない

### useEffect

`useEffect` は **React の外にあるもの（ブラウザAPI・DOM・サードパーティのライブラリ）と同期するためだけ** に使う。
React の中だけで完結する処理（props・state から値を作る、ユーザー操作に応じて処理する）に `useEffect` を使うことを禁止する。
余計な再レンダリングが起きるうえ、「どの操作で何が起きるか」がコードから追いにくくなるため。

#### 禁止するパターン

| # | やりたいこと | ❌ 禁止 | ✅ 代わりに |
|---|---|---|---|
| 1 | props・stateから値を作る | `useEffect` で別の state に `setState` する | レンダリング中に計算する（重ければ `useMemo`） |
| 2 | propsが変わったら state をすべて初期化する | `useEffect` で `setState` する | 親で `key` を変えて、コンポーネントを作り直す |
| 3 | ユーザー操作に応じて処理する（送信・通知・画面遷移） | state を変えて、`useEffect` でそれを拾って処理する | イベントハンドラの中で直接処理する |
| 4 | state の変化に連動して別の state を変える | `useEffect` を連鎖させる（A → effect → B → effect → C） | イベントハンドラの中で、次の state をまとめて計算して更新する |
| 5 | 親に値の変化を知らせる | `useEffect` の中で `onChange` 等の props を呼ぶ | 値を変えたイベントハンドラの中で一緒に呼ぶ |
| 6 | APIからデータを取得する | `useEffect` の中で `fetch` して `setState` する | TanStack Query（`useQuery`）か Server Component |
| 7 | 外部のストア・ブラウザの状態を読む | `useEffect` で購読して `setState` する | `useSyncExternalStore` |

```tsx
// ❌ 1: 派生値を state に同期している
const [lines, setLines] = useState<LineInput[]>([]);
const [debitTotal, setDebitTotal] = useState(0);
useEffect(() => {
  setDebitTotal(sumOf(lines, "debit"));
}, [lines]);

// ✅ レンダリング中に計算する
const debitTotal = sumOf(lines, "debit");
```

```tsx
// ❌ 2: props が変わったら useEffect で入力をリセットしている（一瞬古い値で描画される）
function PostForm({ post }: PostFormProps) {
  const [title, setTitle] = useState(post.title);
  useEffect(() => {
    setTitle(post.title);
  }, [post]);
}

// ✅ 親で key を変えて作り直す
<PostForm key={post.id} post={post} />
```

```tsx
// ❌ 3: 送信をフラグの state 経由で useEffect に任せている
const [submitted, setSubmitted] = useState(false);
useEffect(() => {
  if (submitted) {
    createMutation.mutate(input);
    showToast("保存しました");
  }
}, [submitted]);

// ✅ イベントハンドラで直接処理する
function handleSubmit() {
  createMutation.mutate(input, { onSuccess: () => showToast("保存しました") });
}
```

```tsx
// ❌ 5: 親への通知を useEffect で行っている（親の再レンダリングが1回余計に走る）
useEffect(() => {
  onChange(value);
}, [value]);

// ✅ 値を変えたイベントハンドラで一緒に呼ぶ
function handleChange(next: string) {
  setValue(next);
  onChange(next);
}
```

```tsx
// ❌ 6: useEffect で fetch している（キャッシュ・再取得・競合・エラー処理を自前で書くことになる）
const [accounts, setAccounts] = useState<ChartOfAccount[]>([]);
useEffect(() => {
  fetchChartOfAccounts().then(setAccounts);
}, []);

// ✅ TanStack Query を使う
const { data: accounts = [] } = useChartOfAccounts();
```

#### useEffect を使ってよい場合

- `window` / `document` へのイベントリスナーの登録（`useSession` の `portal:session-change` 等）
- サードパーティのライブラリのインスタンスと同期する（Tiptap のエディタに外から渡された内容を反映する等）
- IntersectionObserver 等のブラウザAPIの結果に応じて処理する（無限スクロールで `fetchNextPage` を呼ぶ等）
- フォーカス・スクロール位置のような、DOM を直接操作する処理
- サーバーとクライアントで値が変わるもの（現在時刻等）を、ハイドレーションの不一致を避けるためにマウント後に設定する

#### useEffect を書くときのルール

- **なぜ useEffect が必要なのか（何と同期しているのか）をコメントに書く**。上の「使ってよい場合」のどれに当たるのかを説明できないなら、禁止パターンになっていないかを見直す
- 登録したイベントリスナー・タイマー・購読は、必ずクリーンアップ関数で解除する
- 非同期処理の結果を反映する場合は、アンマウント・再実行のあとに反映しないよう `ignore` フラグを使う（`app/demo-portal/_lib/useSession.ts` を参照）
- 依存配列をごまかさない。`react-hooks/exhaustive-deps` や `react-hooks/set-state-in-effect` を `eslint-disable` で止めるのは、上の「使ってよい場合」に当たり、理由をコメントに書いた場合だけにする（`eslint-disable` 全般について同様）
- コードレビューで新しく `useEffect` が追加されていたら、禁止パターンに当たらないかを必ず確認する

#### Lint による検出

`eslint-plugin-react-hooks` v7（`eslint-config-next` に含まれる）の以下のルールを error にしている。

- `react-hooks/set-state-in-effect`: `useEffect` の中で同期的に `setState` する（禁止パターン 1・2・4 の多く）を検出する
- `react-hooks/exhaustive-deps`: 依存配列の漏れを検出する

禁止パターン 3・5・6 のように、Lint では検出できないものはコードレビューで確認する。

### Server Component / Client Component

- `page.tsx` / `layout.tsx` は Server Component のままにし、状態やイベントが必要な部分だけを `_components/` の Client Component（`"use client"`）に切り出す
- `"use client"` を付けたファイルから import したものはすべてクライアント側に含まれる。Server Component でしか使わない処理（`getForm` 等）を Client Component から import しない
- 動的ルートの `params` は `Promise` なので `await` する。型は Next.js が生成する `PageProps<"/path/[id]">` / `LayoutProps<"/path">` を使う
- `generateMetadata` とページ本体で同じデータを取得する場合は、React の `cache` でまとめる（`form-builder/_lib/getForm.ts` を参照）。存在しないIDは `notFound()` にする

## データ取得（Go API）

### APIクライアント（`_lib/api.ts`）

- `fetch` の呼び出しは feature ごとの `_lib/api.ts` に集め、コンポーネントから直接 `fetch` しない
- ベースURLは `NEXT_PUBLIC_API_BASE_URL` から取得する（URLを直書きしない）
- `response.ok` を必ず確認し、失敗時はAPIの `{"error": "..."}` を優先したメッセージで `Error` を投げる。APIがメッセージを返さない場合に備えて、日本語のフォールバックメッセージを渡す
- 書き込み系（`POST` / `PUT` / `DELETE`）の `fetch` には `credentials: "include"` を付ける（セッションCookieを送るため）
- 戻り値の型を明示し、APIのレスポンスの型は `_lib/types.ts` に置く

### TanStack Query（`_lib/useXxx.ts`）

- Client Component でのデータ取得・更新は TanStack Query を使う。取得結果を自前の state にコピーしない
  ```ts
  // ✅
  const { data: accounts = [] } = useChartOfAccounts();
  // ❌
  const [accounts, setAccounts] = useState([]);
  useEffect(() => { fetchChartOfAccounts().then(setAccounts); }, []);
  ```
- クエリキーは先頭を feature-slug にした定数にする（例: `["simple-ledger", "transactions"]`）。feature 間でキャッシュが衝突しないようにするため
- 更新（`useMutation`）が成功したら、関係するクエリを `invalidateQueries` で取得し直す
- `QueryClient` は feature の `_components/Providers.tsx` で `useState(() => new QueryClient())` として作る（レンダリングのたびに作り直さないため）

## 入力の検証（zod）

- フォームの入力値の検証スキーマは feature の `_lib/schema.ts` に zod で定義し、エラーメッセージは画面にそのまま出せる日本語にする
- **検証の正は Go API 側**。フロント側の検証は入力の事前チェックで、文字数上限等の制約値を変える場合は Go 側と必ず揃える
- 送信時は `safeParse` を使い、失敗時は例外ではなくエラーメッセージを画面に出す

## スタイリング（Tailwind CSS / `components/ui`）

- ボタン・入力欄・カード・モーダル・表のような繰り返し使う部品は `components/ui` のコンポーネントを使う。feature の中で同じ見た目を Tailwind のクラスで作り直さない
  ```tsx
  // ✅
  <Button variant="secondary" onClick={onClose}>キャンセル</Button>
  // ❌
  <button className="rounded-lg border border-border px-4 py-2.5 text-sm font-semibold">キャンセル</button>
  ```
- ページ内の配置（`flex` / `grid` / `gap` / 余白等）は、Tailwind のクラスを直接書いてよい
- 色は `bg-foreground` / `text-muted-foreground` / `bg-danger` のようなテーマのトークンを使い、`bg-gray-500` や `#xxxxxx` を直書きしない（ダークモード等でテーマを変えたときに追従させるため）
- クラス名を条件で切り替える・外から上書きできるようにする場合は `cn()`（`clsx` + `tailwind-merge`）を使い、文字列の連結にしない
- `components/ui` に部品を追加・変更するときは、ほかの feature で使われている箇所への影響を確認する

## アクセシビリティ

- クリックできる要素は `<button>` / `<a>`（`Link`）にする。`<div onClick>` にしない
- 入力欄には `<label>` を関連付ける。アイコンだけのボタンには `aria-label` を付ける
- 権限がなく押せないボタンは `disabled` ではなく、`Button` の `blockedReason` を使う（`ARCHITECTURE.md` の「アクセス管理（RBAC）」を参照）。理由をフォーカス・読み上げでも伝えるため
- モーダルは `components/ui` の `Modal` を使う（`role="dialog"`・Escキーで閉じる処理が組み込まれている）
- ドラッグ＆ドロップのようなマウス前提の操作には、キーボードで同じことをする手段を用意する

## feature の独立性

- feature から別の feature のファイルを import しない。複数の feature で使う処理は `components/ui`（見た目）か `app/demo-portal/_lib`（セッション等）に移す
- 権限チェックには `useWriteBlockedReason(<feature-slug>, <操作名>)` を使い、feature ごとに独自の認証・権限の仕組みを作らない

## 命名

| 対象 | 規則 | 例 |
|---|---|---|
| コンポーネント・そのファイル | PascalCase | `TransactionFormModal.tsx` |
| Hooks・そのファイル | `use` + camelCase | `useTransactions.ts` |
| そのほかの関数・変数・ファイル | camelCase | `getForm.ts`, `debitTotal` |
| モジュール全体の定数 | UPPER_SNAKE_CASE | `PAGE_SIZE`, `TRANSACTIONS_KEY` |
| イベントハンドラ | 内部の関数は `handleXxx`、propsは `onXxx` | `handleSubmit`, `onClose` |
| 真偽値 | `is` / `has` / `can` 等で始める | `isPending`, `hasNextPage` |

## コメント

- コメントは日本語で書き、**なぜそうしているのか**（理由・制約・避けている問題）を書く。コードを読めば分かる「何をしているか」は書かない
- 仕様上の注意点（ハイドレーションの不一致を避けるため、等）は、コメントに残して次に触る人が同じ問題を踏まないようにする
