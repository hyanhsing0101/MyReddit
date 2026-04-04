"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import {
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiGetBoardBySlug,
  type BoardItem,
} from "@/lib/api";

export default function BoardDetailPage() {
  const params = useParams();
  const router = useRouter();
  const slugRaw = params.slug;
  const slug =
    typeof slugRaw === "string"
      ? slugRaw
      : Array.isArray(slugRaw)
        ? slugRaw[0] ?? ""
        : "";

  const [board, setBoard] = useState<BoardItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!slug) {
      setLoading(false);
      setError("无效的板块地址");
      return;
    }
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      try {
        const body = await apiGetBoardBySlug(slug);
        if (cancelled) return;
        if (body.code !== API_SUCCESS_CODE || !body.data) {
          setError(apiErrorMessage(body));
          setBoard(null);
          return;
        }
        setBoard(body.data);
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "加载失败");
          setBoard(null);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [slug]);

  return (
    <div className="mx-auto max-w-2xl px-4 py-10">
      <button
        type="button"
        onClick={() => router.push("/boards")}
        className="mb-6 text-sm text-zinc-500 underline"
      >
        ← 返回板块列表
      </button>

      {loading ? (
        <p className="text-sm text-zinc-500">加载中…</p>
      ) : error ? (
        <div className="rounded-lg border border-zinc-200 p-6 dark:border-zinc-800">
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
          <Link href="/boards" className="mt-4 inline-block text-sm underline">
            回列表
          </Link>
        </div>
      ) : board ? (
        <article>
          <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
            {board.name}
          </h1>
          <p className="mt-2 text-sm text-zinc-500">
            slug：<code className="rounded bg-zinc-200 px-1 dark:bg-zinc-800">{board.slug}</code>
            {board.is_system_sink ? (
              <span className="ml-2 text-amber-700 dark:text-amber-400">
                （系统板）
              </span>
            ) : null}
          </p>
          {board.description ? (
            <p className="mt-6 whitespace-pre-wrap text-sm text-zinc-700 dark:text-zinc-300">
              {board.description}
            </p>
          ) : (
            <p className="mt-6 text-sm text-zinc-500">暂无描述</p>
          )}
          <p className="mt-8 text-xs text-zinc-500">
            帖子与板块的关联后续在此展示。
          </p>
          <p className="mt-2">
            <Link href="/" className="text-sm text-zinc-500 underline">
              首页
            </Link>
          </p>
        </article>
      ) : null}
    </div>
  );
}
