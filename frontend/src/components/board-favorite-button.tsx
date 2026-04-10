"use client";

import { useState } from "react";
import {
  API_SUCCESS_CODE,
  apiAddBoardFavorite,
  apiErrorMessage,
  apiRemoveBoardFavorite,
} from "@/lib/api";

type Props = {
  boardId: number;
  isSystemSink: boolean;
  /** public 板不提供订阅 */
  visibility: string;
  isFavorited: boolean;
  accessToken: string | null | undefined;
  onUpdated: (next: boolean) => void;
  className?: string;
};

export function BoardFavoriteButton({
  boardId,
  isSystemSink,
  visibility,
  isFavorited,
  accessToken,
  onUpdated,
  className = "",
}: Props) {
  const [busy, setBusy] = useState(false);
  const token = accessToken ?? "";
  const vis = visibility || "public";
  if (isSystemSink || vis === "public" || !token) return null;

  async function toggle(e: React.MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    if (busy) return;
    setBusy(true);
    try {
      const res = isFavorited
        ? await apiRemoveBoardFavorite(token, boardId)
        : await apiAddBoardFavorite(token, boardId);
      if (res.code !== API_SUCCESS_CODE) {
        window.alert(apiErrorMessage(res));
        return;
      }
      onUpdated(!isFavorited);
    } catch (err) {
      window.alert(err instanceof Error ? err.message : "操作失败");
    } finally {
      setBusy(false);
    }
  }

  return (
    <button
      type="button"
      disabled={busy}
      onClick={toggle}
      title={isFavorited ? "取消订阅" : "订阅板块（私有板）"}
      className={`shrink-0 rounded-lg border border-zinc-300 px-2 py-1 text-sm disabled:opacity-50 dark:border-zinc-600 ${className}`}
    >
      {busy ? "…" : isFavorited ? "★ 已订阅" : "☆ 订阅"}
    </button>
  );
}
