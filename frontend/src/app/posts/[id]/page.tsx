"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import {
  API_POST_NOT_EXIST_CODE,
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiGetPost,
  type PostItem,
} from "@/lib/api";

export default function PostDetailPage() {
  const params = useParams();
  const router = useRouter();
  const idParam = params.id;
  const id =
    typeof idParam === "string"
      ? Number.parseInt(idParam, 10)
      : Array.isArray(idParam)
        ? Number.parseInt(idParam[0] ?? "", 10)
        : NaN;

  const [post, setPost] = useState<PostItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    if (!Number.isFinite(id) || id < 1) {
      setLoading(false);
      setError("无效的帖子 ID");
      return;
    }

    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      setNotFound(false);
      try {
        const body = await apiGetPost(id);
        if (cancelled) return;
        if (body.code === API_POST_NOT_EXIST_CODE) {
          setNotFound(true);
          setPost(null);
          return;
        }
        if (body.code !== API_SUCCESS_CODE || !body.data) {
          setError(apiErrorMessage(body));
          setPost(null);
          return;
        }
        setPost(body.data);
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "加载失败");
          setPost(null);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [id]);

  return (
    <div className="mx-auto max-w-2xl px-4 py-10">
      <button
        type="button"
        onClick={() => router.push("/")}
        className="mb-6 text-sm text-zinc-500 underline"
      >
        ← 返回首页
      </button>

      {loading ? (
        <p className="text-sm text-zinc-500">加载中…</p>
      ) : notFound ? (
        <div className="rounded-lg border border-zinc-200 p-6 dark:border-zinc-800">
          <p className="text-zinc-700 dark:text-zinc-300">帖子不存在或已删除。</p>
          <Link href="/" className="mt-4 inline-block text-sm underline">
            回首页
          </Link>
        </div>
      ) : error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : post ? (
        <article>
          <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
            {post.title}
          </h1>
          <p className="mt-3 text-xs text-zinc-500">
            {post.author_id != null
              ? `作者 ID ${post.author_id}`
              : "无主帖"}{" "}
            ·{" "}
            {new Date(post.create_time).toLocaleString("zh-CN", {
              year: "numeric",
              month: "numeric",
              day: "numeric",
              hour: "2-digit",
              minute: "2-digit",
            })}
          </p>
          <div className="mt-8 whitespace-pre-wrap break-words text-sm leading-relaxed text-zinc-800 dark:text-zinc-200">
            {post.content}
          </div>
        </article>
      ) : null}
    </div>
  );
}
