"use client";

import { useState } from "react";
import Link from "next/link";
import { Badge, Button, Card, Select, Table, TableCell, TableHeaderCell, TableRow, cn } from "@/components/ui";
import type { PermissionLevel } from "../../_lib/auth";
import { useSession } from "../../_lib/useSession";
import { ApiError } from "../_lib/api";
import type { Role, RolesResponse } from "../_lib/api";
import { isSamePermissions, levelLabels, resourceLabel, selectableLevels } from "../_lib/permissions";
import { useRoles, useUpdateRolePermissions } from "../_lib/useRoles";

type Drafts = Record<number, Record<string, PermissionLevel>>;

function toDrafts(roles: Role[]): Drafts {
  return Object.fromEntries(roles.map((role) => [role.id, { ...role.permissions }]));
}

// ロール×リソースの権限表。access-control の write 権限があれば、各セルの権限を変えて保存できる。
// 変更したロールだけをロールごとに PUT /api/access-control/roles/:id/permissions で保存する。
export function RolePermissionsEditor() {
  const { data, error } = useRoles();
  const session = useSession();

  if (error) {
    return <LoadError error={error} />;
  }
  // 権限の確認が終わる前に表を出すと、編集できる表示に途中で切り替わるため、両方そろうまで待つ
  if (!data || session === undefined) {
    return <p className="text-sm text-muted-foreground">読み込み中...</p>;
  }

  return <RolePermissionsTable data={data} canWrite={session?.permissions["access-control"] === "write"} />;
}

type RolePermissionsTableProps = {
  data: RolesResponse;
  canWrite: boolean;
};

// drafts の初期値を取得済みのロールから作るため、取得が終わってから描画するコンポーネントに分けている
function RolePermissionsTable({ data, canWrite }: RolePermissionsTableProps) {
  const [drafts, setDrafts] = useState<Drafts>(() => toDrafts(data.roles));
  const [message, setMessage] = useState<{ variant: "success" | "error"; text: string } | null>(null);
  const updateMutation = useUpdateRolePermissions();
  const isSaving = updateMutation.isPending;

  const changedRoles = data.roles.filter((role) => !isSamePermissions(drafts[role.id] ?? {}, role.permissions));

  function handleChange(roleId: number, resource: string, level: PermissionLevel) {
    setMessage(null);
    setDrafts((current) => ({ ...current, [roleId]: { ...current[roleId], [resource]: level } }));
  }

  function handleReset() {
    setMessage(null);
    setDrafts(toDrafts(data.roles));
  }

  function handleSave() {
    setMessage(null);
    updateMutation.mutate(
      changedRoles.map((role) => ({ roleId: role.id, permissions: drafts[role.id] })),
      {
        onSuccess: () => setMessage({ variant: "success", text: "権限を保存しました" }),
        onError: (error) => setMessage({ variant: "error", text: error.message || "権限の保存に失敗しました" }),
      },
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <p className="text-sm text-muted-foreground">
        「編集」は「閲覧」を含みます。ユーザーの権限は、割り当てられたすべてのロールとゲストロールのうち、リソースごとに最も強いものになります。
        {canWrite ? null : "閲覧のみの権限のため、変更はできません。"}
      </p>

      <Card className="p-0">
        <Table>
          <thead>
            <tr>
              <TableHeaderCell scope="col">リソース</TableHeaderCell>
              {data.roles.map((role) => (
                <TableHeaderCell key={role.id} scope="col">
                  <span className="flex flex-col gap-1.5">
                    <span className="flex flex-wrap items-center gap-2 font-semibold text-foreground">
                      {role.name}
                      {role.editable ? null : (
                        <Badge variant="neutral" dot={false}>
                          変更不可
                        </Badge>
                      )}
                    </span>
                    {role.description ? <span className="text-xs font-normal">{role.description}</span> : null}
                  </span>
                </TableHeaderCell>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.resources.map((definition) => (
              <TableRow key={definition.resource}>
                <th scope="row" className="px-5 py-4 text-left align-middle font-semibold">
                  {resourceLabel(definition.resource)}
                </th>
                {data.roles.map((role) => {
                  const levels = selectableLevels(role, definition);
                  const value = drafts[role.id]?.[definition.resource] ?? "none";
                  const changed = value !== role.permissions[definition.resource];
                  const editable = canWrite && role.editable && levels.length > 1;
                  return (
                    <TableCell key={role.id} className={cn("py-4 align-middle", changed ? "bg-accent-bg" : undefined)}>
                      {editable ? (
                        <Select
                          aria-label={`${role.name}の「${resourceLabel(definition.resource)}」の権限`}
                          value={value}
                          onChange={(e) => handleChange(role.id, definition.resource, e.target.value as PermissionLevel)}
                          disabled={isSaving}
                          className="min-w-24"
                        >
                          {levels.map((level) => (
                            <option key={level} value={level}>
                              {levelLabels[level]}
                            </option>
                          ))}
                        </Select>
                      ) : (
                        <LevelLabel level={value} />
                      )}
                    </TableCell>
                  );
                })}
              </TableRow>
            ))}
          </tbody>
        </Table>
      </Card>

      {canWrite ? (
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-end">
          {message ? <SaveMessage {...message} /> : null}
          {!message && changedRoles.length > 0 ? (
            <p className="text-sm text-muted-foreground">{changedRoles.length}件のロールに未保存の変更があります</p>
          ) : null}
          <Button
            variant="secondary"
            onClick={handleReset}
            disabled={isSaving || changedRoles.length === 0}
            className="sm:w-auto"
          >
            元に戻す
          </Button>
          <Button onClick={handleSave} disabled={isSaving || changedRoles.length === 0} className="sm:w-auto">
            {isSaving ? "保存中..." : "変更を保存"}
          </Button>
        </div>
      ) : message ? (
        <SaveMessage {...message} />
      ) : null}
    </div>
  );
}

function LevelLabel({ level }: { level: PermissionLevel }) {
  const variant = level === "write" ? "accent" : level === "read" ? "success" : "neutral";
  return (
    <Badge variant={variant} dot={false}>
      {levelLabels[level]}
    </Badge>
  );
}

function SaveMessage({ variant, text }: { variant: "success" | "error"; text: string }) {
  return (
    <p role={variant === "error" ? "alert" : "status"} className={cn("text-sm", variant === "error" ? "text-danger" : "text-success")}>
      {text}
    </p>
  );
}

function LoadError({ error }: { error: Error }) {
  if (error instanceof ApiError && error.status === 401) {
    return (
      <Card>
        <p className="text-sm">アクセス管理を見るにはログインが必要です。</p>
        <Link href="/demo-portal/login?redirect=/demo-portal/access-control" className="text-sm font-semibold text-accent hover:underline">
          ログイン画面へ
        </Link>
      </Card>
    );
  }
  if (error instanceof ApiError && error.status === 403) {
    return (
      <Card>
        <p className="text-sm">アクセス管理を見る権限がありません。</p>
      </Card>
    );
  }
  return (
    <p role="alert" className="text-sm text-danger">
      {error.message}
    </p>
  );
}
