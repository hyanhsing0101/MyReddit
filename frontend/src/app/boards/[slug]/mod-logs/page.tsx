"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import {
  API_BOARD_NOT_EXIST_CODE,
  API_FORBIDDEN_CODE,
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiGetBoardBySlug,
  apiListBoardModerationLogs,
  type BoardItem,
  type ModerationAction,
  type ModerationLogItem,
} from "@/lib/api";
import { getAccessToken } from "@/lib/auth-storage";

const ACTION_OPTIONS: Array<{ id: "" | ModerationAction; label: string }> = [
  { id: "", label: "全部动作" },
  { id: "seal_post", label: "封帖" },
  { id: "unseal_post", label: "解封" },
  { id: "delete_post", label: "软删帖子" },
  { id: "restore_post", label: "恢复帖子" },
  { id: "handle_post_appeal", label: "处理申诉" },
  { id: "lock_post_comments", label: "锁评" },
  { id: "unlock_post_comments", label: "解锁评论" },
  { id: "pin_post", label: "置顶" },
  { id: "unpin_post", label: "取消置顶" },
  { id: "update_post_report_status", label: "处理帖子举报" },
  { id: "update_comment_report_status", label: "处理评论举报" },
  { id: "upsert_board_moderator", label: "新增/更新版主" },
  { id: "update_board_moderator_role", label: "更新版主角色" },
  { id: "remove_board_moderator", label: "移除版主" },
];

const ACTION_LABEL: Record<ModerationAction, string> = {
  seal_post: "封帖",
  unseal_post: "解封",
  delete_post: "软删帖子",
  restore_post: "恢复帖子",
  handle_post_appeal: "处理申诉",
  lock_post_comments: "锁评",
  unlock_post_comments: "解锁评论",
  pin_post: "置顶",
  unpin_post: "取消置顶",
  update_post_report_status: "处理帖子举报",
  update_comment_report_status: "处理评论举报",
  upsert_board_moderator: "新增/更新版主",
  update_board_moderator_role: "更新版主角色",
  remove_board_moderator: "移除版主",
};

export default function BoardModerationLogsPage() {
  const params = useParams();
  const router = useRouter();
  const slugRaw = params.slug;
  const slug =
    typeof slugRaw === "string"
      ? slugRaw
      : Array.isArray(slugRaw)
        ? (slugRaw[0] ?? "")
        : "";

  const [board, setBoard] = useState<BoardItem | null>(null);
  const [rows, setRows] = useState<ModerationLogItem[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [pageSize] = useState(20);
  const [action, setAction] = useState<"" | ModerationAction>("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!slug) {
      setError("无效的板块地址");
      setLoading(false);
      return;
    }
    const token = getAccessToken();
    if (!token) {
      setError("请先登录");
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const b = await apiGetBoardBySlug(slug, token);
      if (b.code === API_BOARD_NOT_EXIST_CODE) {
        setError("板块不存在或无权访问");
        setBoard(null);
        setRows([]);
        return;
      }
      if (b.code !== API_SUCCESS_CODE || !b.data) {
        setError(apiErrorMessage(b));
        setBoard(null);
        setRows([]);
        return;
      }
      setBoard(b.data);
      const list = await apiListBoardModerationLogs(token, b.data.id, page, pageSize, action);
      if (list.code === API_FORBIDDEN_CODE) {
        setError("你不是该板块版主或站点管理员");
        setRows([]);
        setTotal(0);
        return;
      }
      if (list.code !== API_SUCCESS_CODE || !list.data) {
        setError(apiErrorMessage(list));
        setRows([]);
        setTotal(0);
        return;
      }
      setRows(Array.isArray(list.data.list) ? list.data.list : []);
      setTotal(list.data.total ?? 0);
      setPage(list.data.page ?? page);
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载失败");
      setRows([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [slug, page, pageSize, action]);

  useEffect(() => {
    void load();
  }, [load]);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  return (
    <div className="mx-auto max-w-3xl px-4 py-10">
      <button
        type="button"
        onClick={() => router.push(`/boards/${encodeURIComponent(slug)}`)}
        className="mb-6 text-sm text-zinc-500 underline"
      >
        ← 返回板块详情
      </button>

      <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
        治理日志
      </h1>
      {board ? (
        <p className="mt-1 text-sm text-zinc-500">
          板块：{board.name} /{board.slug}
        </p>
      ) : null}

      <div className="mt-4 flex flex-wrap items-center gap-2">
        {ACTION_OPTIONS.map((opt) => (
          <button
            key={opt.id || "all"}
            type="button"
            onClick={() => {
              setAction(opt.id);
              setPage(1);
            }}
            className={
              action === opt.id
                ? "rounded-lg border border-zinc-900 bg-zinc-900 px-3 py-1 text-xs text-white dark:border-zinc-100 dark:bg-zinc-100 dark:text-zinc-900"
                : "rounded-lg border border-zinc-300 px-3 py-1 text-xs dark:border-zinc-600"
            }
          >
            {opt.label}
          </button>
        ))}
      </div>

      {loading ? <p className="mt-6 text-sm text-zinc-500">加载中…</p> : null}
      {error ? (
        <p className="mt-4 text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : null}

      {!loading && !error ? (
        <section className="mt-6 rounded-xl border border-zinc-200 dark:border-zinc-800">
          <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
            {rows.length === 0 ? (
              <p className="px-4 py-6 text-sm text-zinc-500">暂无日志</p>
            ) : (
              rows.map((r) => (
                <article key={r.id} className="px-4 py-4">
                  <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                    {ACTION_LABEL[r.action]} · 目标 {r.target_type}:{r.target_id}
                  </p>
                  <p className="mt-1 text-xs text-zinc-500">
                    操作者：{r.operator_username || r.operator_id} ·{" "}
                    {new Date(r.create_time).toLocaleString("zh-CN")}
                  </p>
                  {r.description ? (
                    <p className="mt-2 whitespace-pre-wrap text-sm text-zinc-700 dark:text-zinc-300">
                      {r.description}
                    </p>
                  ) : null}
                </article>
              ))
            )}
          </div>
          {!loading && rows.length > 0 && total > pageSize ? (
            <div className="flex items-center justify-between border-t border-zinc-200 px-4 py-3 text-sm dark:border-zinc-800">
              <span className="text-zinc-500">
                第 {page} / {totalPages} 页 · 共 {total} 条
              </span>
              <div className="flex gap-2">
                <button
                  type="button"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                >
                  上一页
                </button>
                <button
                  type="button"
                  disabled={page >= totalPages}
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                >
                  下一页
                </button>
              </div>
            </div>
          ) : null}
        </section>
      ) : null}

      <p className="mt-6">
        <Link
          href={`/boards/${encodeURIComponent(slug)}/moderation`}
          className="mr-4 text-sm text-zinc-500 underline"
        >
          治理工作台
        </Link>
        <Link href="/boards" className="text-sm text-zinc-500 underline">
          板块列表
        </Link>
      </p>
    </div>
  );
}
