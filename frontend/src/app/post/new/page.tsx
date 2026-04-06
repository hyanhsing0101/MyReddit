"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import {
  API_SUCCESS_CODE,
  apiCreatePost,
  apiErrorMessage,
  apiListBoards,
  apiListTags,
  tagDisplayLabel,
  type BoardItem,
  type TagItem,
} from "@/lib/api";
import { getAccessToken } from "@/lib/auth-storage";

function NewPostForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const boardIdFromUrl = searchParams.get("board_id");

  const [boards, setBoards] = useState<BoardItem[]>([]);
  const [boardsError, setBoardsError] = useState<string | null>(null);
  const [boardsLoading, setBoardsLoading] = useState(true);
  const [boardId, setBoardId] = useState<number | "">("");
  const [tags, setTags] = useState<TagItem[]>([]);
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([]);
  const [tagsLoading, setTagsLoading] = useState(true);
  const [tagsError, setTagsError] = useState<string | null>(null);

  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [ok, setOk] = useState(false);
  const [loading, setLoading] = useState(false);
  const [hasToken, setHasToken] = useState<boolean | null>(null);

  useEffect(() => {
    setHasToken(!!getAccessToken());
  }, []);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setTagsLoading(true);
      setTagsError(null);
      try {
        const body = await apiListTags(1, 100);
        if (cancelled) return;
        if (body.code !== API_SUCCESS_CODE || !body.data) {
          setTagsError(apiErrorMessage(body));
          setTags([]);
          return;
        }
        setTags(body.data.list);
      } catch (e) {
        if (!cancelled) {
          setTagsError(e instanceof Error ? e.message : "加载标签失败");
          setTags([]);
        }
      } finally {
        if (!cancelled) setTagsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setBoardsLoading(true);
      setBoardsError(null);
      try {
        const body = await apiListBoards(1, 100, false);
        if (cancelled) return;
        if (body.code !== API_SUCCESS_CODE || !body.data) {
          setBoardsError(apiErrorMessage(body));
          setBoards([]);
          return;
        }
        setBoards(body.data.list);
      } catch (e) {
        if (!cancelled) {
          setBoardsError(e instanceof Error ? e.message : "加载板块失败");
          setBoards([]);
        }
      } finally {
        if (!cancelled) setBoardsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (boards.length === 0) return;
    const qid = boardIdFromUrl ? Number.parseInt(boardIdFromUrl, 10) : NaN;
    if (Number.isFinite(qid) && boards.some((b) => b.id === qid)) {
      setBoardId(qid);
      return;
    }
    const general = boards.find((b) => b.slug === "general");
    setBoardId(general?.id ?? boards[0].id);
  }, [boards, boardIdFromUrl]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setOk(false);
    const token = getAccessToken();
    if (!token) {
      setError("请先登录");
      return;
    }
    if (boardId === "" || typeof boardId !== "number") {
      setError("请选择板块");
      return;
    }
    setLoading(true);
    try {
      const body = await apiCreatePost(token, {
        board_id: boardId,
        tag_ids: selectedTagIds,
        title,
        content,
      });
      if (body.code !== API_SUCCESS_CODE) {
        setError(apiErrorMessage(body));
        return;
      }
      setOk(true);
      setTitle("");
      setContent("");
      setSelectedTagIds([]);
    } catch (err) {
      setError(err instanceof Error ? err.message : "发布失败");
    } finally {
      setLoading(false);
    }
  }

  if (hasToken === null) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-16 text-sm text-zinc-500">
        加载中…
      </div>
    );
  }

  if (!hasToken) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-16">
        <p className="text-zinc-700 dark:text-zinc-300">发帖需要先登录。</p>
        <Link
          href="/login"
          className="mt-4 inline-block text-sm text-zinc-900 underline dark:text-zinc-100"
        >
          去登录
        </Link>
        <p className="mt-6">
          <Link href="/" className="text-sm text-zinc-500 underline">
            返回首页
          </Link>
        </p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-16">
      <div className="mb-8 flex items-center justify-between gap-4">
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
          发帖
        </h1>
        <button
          type="button"
          onClick={() => router.push("/")}
          className="text-sm text-zinc-500 underline"
        >
          返回首页
        </button>
      </div>

      {boardsLoading ? (
        <p className="mb-4 text-sm text-zinc-500">加载板块列表…</p>
      ) : boardsError ? (
        <p className="mb-4 text-sm text-red-600 dark:text-red-400">
          {boardsError}
        </p>
      ) : boards.length === 0 ? (
        <p className="mb-4 text-sm text-zinc-600 dark:text-zinc-400">
          没有可选板块，请先
          <Link href="/boards/new" className="mx-1 underline">
            创建板块
          </Link>
          。
        </p>
      ) : null}
      {tagsLoading ? (
        <p className="mb-4 text-sm text-zinc-500">加载标签列表…</p>
      ) : tagsError ? (
        <p className="mb-4 text-sm text-red-600 dark:text-red-400">
          {tagsError}
        </p>
      ) : null}

      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-zinc-600 dark:text-zinc-400">板块</span>
          <select
            className="rounded-lg border border-zinc-300 bg-white px-3 py-2 dark:border-zinc-600 dark:bg-zinc-900"
            value={boardId === "" ? "" : String(boardId)}
            onChange={(e) => {
              const v = e.target.value;
              setBoardId(v === "" ? "" : Number.parseInt(v, 10));
            }}
            required
            disabled={boards.length === 0 || boardsLoading}
          >
            {boards.map((b) => (
              <option key={b.id} value={b.id}>
                {b.name} ({b.slug})
                {b.is_system_sink ? " · 系统" : ""}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-zinc-600 dark:text-zinc-400">标题</span>
          <input
            className="rounded-lg border border-zinc-300 bg-white px-3 py-2 dark:border-zinc-600 dark:bg-zinc-900"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
            maxLength={255}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-zinc-600 dark:text-zinc-400">正文（纯文本）</span>
          <textarea
            className="min-h-[240px] rounded-lg border border-zinc-300 bg-white px-3 py-2 font-mono text-sm dark:border-zinc-600 dark:bg-zinc-900"
            value={content}
            onChange={(e) => setContent(e.target.value)}
            required
            rows={12}
          />
        </label>
        <fieldset className="flex flex-col gap-2 text-sm">
          <legend className="text-zinc-600 dark:text-zinc-400">标签（最多 5 个）</legend>
          <div className="flex flex-wrap gap-2">
            {tags.length === 0 ? (
              <span className="text-xs text-zinc-500">暂无可选标签</span>
            ) : (
              tags.map((tag) => {
                const checked = selectedTagIds.includes(tag.id);
                const disabled = !checked && selectedTagIds.length >= 5;
                return (
                  <label
                    key={tag.id}
                    className="inline-flex items-center gap-1 rounded border border-zinc-300 px-2 py-1 dark:border-zinc-600"
                  >
                    <input
                      type="checkbox"
                      checked={checked}
                      disabled={disabled}
                      onChange={(e) => {
                        if (e.target.checked) {
                          setSelectedTagIds((prev) => [...prev, tag.id]);
                        } else {
                          setSelectedTagIds((prev) =>
                            prev.filter((id) => id !== tag.id),
                          );
                        }
                      }}
                    />
                    <span>#{tagDisplayLabel(tag)}</span>
                  </label>
                );
              })
            )}
          </div>
        </fieldset>
        {error ? (
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        ) : null}
        {ok ? (
          <p className="text-sm text-emerald-700 dark:text-emerald-400">
            发布成功。
          </p>
        ) : null}
        <button
          type="submit"
          disabled={loading || boards.length === 0}
          className="rounded-lg bg-zinc-900 py-2.5 text-sm font-medium text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900"
        >
          {loading ? "提交中…" : "发布"}
        </button>
      </form>
    </div>
  );
}

export default function NewPostPage() {
  return (
    <Suspense
      fallback={
        <div className="mx-auto max-w-2xl px-4 py-16 text-sm text-zinc-500">
          加载中…
        </div>
      }
    >
      <NewPostForm />
    </Suspense>
  );
}
