"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchRoles, updateRolePermissions } from "./api";
import type { PermissionLevel } from "../../_lib/auth";
import type { Role, RolesResponse } from "./api";

const ROLES_KEY = ["access-control", "roles"];

export function useRoles() {
  return useQuery({
    queryKey: ROLES_KEY,
    queryFn: fetchRoles,
    // 401（未ログイン）・403（権限なし）は再試行しても変わらず、画面の表示が遅れるだけのため再試行しない
    retry: false,
  });
}

// 変更したロールの権限を、ロールごとに PUT で保存する。
// 途中で失敗した場合も、保存できたロールはキャッシュに反映して表に出す。
export function useUpdateRolePermissions() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (changes: { roleId: number; permissions: Record<string, PermissionLevel> }[]) => {
      for (const { roleId, permissions } of changes) {
        const updated = await updateRolePermissions(roleId, permissions);
        queryClient.setQueryData<RolesResponse>(ROLES_KEY, (current) =>
          current && { ...current, roles: current.roles.map((role: Role) => (role.id === updated.id ? updated : role)) },
        );
      }
    },
  });
}
