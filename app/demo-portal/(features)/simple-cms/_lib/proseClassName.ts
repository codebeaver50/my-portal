// Tailwindの typography プラグインを追加せず、ProseMirrorが出力する要素に
// 直接ユーティリティクラスを当てて最低限の記事スタイルを作る。
export const PROSE_CLASSNAME = [
  "prose-content",
  "flex flex-col gap-4 text-sm leading-relaxed text-foreground",
  "[&_h1]:text-2xl [&_h1]:font-extrabold [&_h1]:tracking-tight",
  "[&_h2]:text-xl [&_h2]:font-bold",
  "[&_h3]:text-lg [&_h3]:font-bold",
  "[&_p]:leading-relaxed",
  "[&_a]:text-accent [&_a]:underline [&_a]:underline-offset-2",
  "[&_strong]:font-bold",
  "[&_ul]:list-disc [&_ul]:pl-5 [&_ol]:list-decimal [&_ol]:pl-5 [&_li]:my-1",
  "[&_blockquote]:border-l-2 [&_blockquote]:border-accent [&_blockquote]:pl-4 [&_blockquote]:text-muted-foreground",
  "[&_pre]:overflow-x-auto [&_pre]:rounded-lg [&_pre]:bg-neutral-bg [&_pre]:p-4 [&_pre]:text-xs",
  "[&_code]:rounded [&_code]:bg-neutral-bg [&_code]:px-1 [&_code]:py-0.5 [&_code]:text-xs",
  "[&_pre_code]:bg-transparent [&_pre_code]:p-0",
  "[&_hr]:border-border",
].join(" ");
