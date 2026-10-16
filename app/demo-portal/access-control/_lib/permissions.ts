import { features } from "@/app/_lib/features";
import type { PermissionLevel } from "../../_lib/auth";
import type { ResourceDefinition, Role } from "./api";

const levelOrder: PermissionLevel[] = ["none", "read", "write"];

export const levelLabels: Record<PermissionLevel, string> = {
  none: "なし",
  read: "閲覧",
  write: "編集",
};

// resourceLabel はリソース（feature-slug または access-control）の表示名を返す。
export function resourceLabel(resource: string): string {
  if (resource === "access-control") return "アクセス管理";
  return features.find((feature) => feature.slug === resource)?.name ?? resource;
}

// selectableLevels はロール×リソースに設定できる権限レベルを弱い順で返す（Go側の検証と揃える）。
// ゲストロールは未ログインの訪問者全員に適用されるため、アクセス管理の権限は付けられない。
export function selectableLevels(role: Role, definition: ResourceDefinition): PermissionLevel[] {
  if (role.code === "guest" && definition.resource === "access-control") return ["none"];
  return levelOrder.slice(0, levelOrder.indexOf(definition.maxLevel) + 1);
}

export function isSamePermissions(a: Record<string, PermissionLevel>, b: Record<string, PermissionLevel>): boolean {
  const keys = new Set([...Object.keys(a), ...Object.keys(b)]);
  return [...keys].every((key) => (a[key] ?? "none") === (b[key] ?? "none"));
}
