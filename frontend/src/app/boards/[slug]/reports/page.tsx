"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import {
  API_BOARD_NOT_EXIST_CODE,
  API_FORBIDDEN_CODE,
  API_POST_REPORT_NOT_EXIST_CODE,
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiGetBoardBySlug,
  apiListBoardReports,
  apiUpdateBoardReportStatus,
  type BoardItem,
  type PostReportItem,
  type PostReportStatus,
} from "@/lib/api";
import { getAccessToken } from "@/lib/auth-storage";

const STATUS_OPTIONS: Array<{ id: "" | PostReportStatus; label: string }> = [
  { id: "", label: "全部" },
  { id: "open", label: "待处理" },
  { id: "in_review", label: "处理中" },
  { id: "resolved", label: "已解决" },
  { id: "rejected", label: "已驳回" },
];

const STATUS_LABEL: Record<PostReportStatus, string> = {
  open: "待处理",
  in_review: "处理中",
  resolved: "已解决",
  rejected: "已驳回",
};

export default function BoardReportsPage() {
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
  const [rows, setRows] = useState<PostReportItem[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [pageSize] = useState(20);
  const [status, setStatus] = useState<"" | PostReportStatus>("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busyReportID, setBusyReportID] = useState<number | null>(null);

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
      const list = await apiListBoardReports(token, b.data.id, page, pageSize, status);
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
  }, [slug, page, pageSize, status]);

  useEffect(() => {
    void load();
  }, [load]);

  async function onChangeStatus(
    reportID: number,
    next: PostReportStatus,
    currentNote: string,
  ) {
    if (!board) return;
    const token = getAccessToken();
    if (!token) {
      setError("请先登录");
      return;
    }
    setBusyReportID(reportID);
    setError(null);
    try {
      const note = window.prompt("可选：填写处理备注（留空可不改）", currentNote) ?? "";
      const res = await apiUpdateBoardReportStatus(token, board.id, reportID, {
        status: next,
        handler_note: note.trim(),
      });
      if (res.code !== API_SUCCESS_CODE) {
        if (res.code === API_POST_REPORT_NOT_EXIST_CODE) {
          setError("举报不存在，列表可能已变化");
          await load();
          return;
        }
        if (res.code === API_FORBIDDEN_CODE) {
          setError("无权处理该举报");
          return;
        }
        setError(apiErrorMessage(res));
        return;
      }
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "操作失败");
    } finally {
      setBusyReportID(null);
    }
  }

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
        举报治理
      </h1>
      {board ? (
        <p className="mt-1 text-sm text-zinc-500">
          板块：{board.name} /{board.slug}
        </p>
      ) : null}

      <div className="mt-4 flex flex-wrap items-center gap-2">
        {STATUS_OPTIONS.map((opt) => (
          <button
            key={opt.id || "all"}
            type="button"
            onClick={() => {
              setStatus(opt.id);
              setPage(1);
            }}
            className={
              status === opt.id
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
              <p className="px-4 py-6 text-sm text-zinc-500">暂无举报</p>
            ) : (
              rows.map((r) => (
                <article key={r.id} className="px-4 py-4">
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                        #{r.id} ·
                        <Link href={`/posts/${r.post_id}`} className="ml-1 underline">
                          {r.post_title || `帖子 ${r.post_id}`}
                        </Link>
                      </p>
                      <p className="mt-1 text-xs text-zinc-500">
                        举报人：{r.reporter_username || r.reporter_id} · 状态：
                        {STATUS_LABEL[r.status]}
                      </p>
                    </div>
                    <select
                      value={r.status}
                      disabled={busyReportID === r.id}
                      onChange={(e) =>
                        void onChangeStatus(
                          r.id,
                          e.target.value as PostReportStatus,
                          r.handler_note ?? "",
                        )
                      }
                      className="rounded border border-zinc-300 px-2 py-1 text-xs dark:border-zinc-600 dark:bg-zinc-900"
                    >
                      <option value="open">待处理</option>
                      <option value="in_review">处理中</option>
                      <option value="resolved">已解决</option>
                      <option value="rejected">已驳回</option>
                    </select>
                  </div>
                  <p className="mt-2 text-sm text-zinc-800 dark:text-zinc-200">
                    原因：{r.reason}
                  </p>
                  {r.detail ? (
                    <p className="mt-1 whitespace-pre-wrap text-sm text-zinc-600 dark:text-zinc-300">
                      {r.detail}
                    </p>
                  ) : null}
                  <p className="mt-2 text-xs text-zinc-500">
                    处理人：{r.handler_username || (r.handler_id ?? "-")} · 更新于{" "}
                    {new Date(r.update_time).toLocaleString("zh-CN")}
                  </p>
                  {r.handler_note ? (
                    <p className="mt-1 whitespace-pre-wrap text-xs text-zinc-500">
                      处理备注：{r.handler_note}
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
