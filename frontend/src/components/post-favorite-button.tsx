"use client";

import { useState } from "react";
import {
  API_SUCCESS_CODE,
  apiAddPostFavorite,
  apiErrorMessage,
  apiRemovePostFavorite,
} from "@/lib/api";

type Props = {
  postId: number;
  isFavorited: boolean;
  accessToken: string | null | undefined;
  onUpdated: (next: boolean) => void;
  className?: string;
};

export function PostFavoriteButton({
  postId,
  isFavorited,
  accessToken,
  onUpdated,
  className = "",
}: Props) {
  const [busy, setBusy] = useState(false);
  const token = accessToken ?? "";
  if (!token) return null;

  async function toggle(e: React.MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    if (busy) return;
    setBusy(true);
    try {
      const res = isFavorited
        ? await apiRemovePostFavorite(token, postId)
        : await apiAddPostFavorite(token, postId);
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
      title={isFavorited ? "取消收藏" : "收藏帖子"}
      className={`shrink-0 rounded-lg border border-zinc-300 px-2 py-1 text-sm disabled:opacity-50 dark:border-zinc-600 ${className}`}
    >
      {busy ? "…" : isFavorited ? "★ 已收藏" : "☆ 收藏"}
    </button>
  );
}
