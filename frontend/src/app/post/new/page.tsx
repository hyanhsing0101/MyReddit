"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { API_SUCCESS_CODE, apiCreatePost, apiErrorMessage } from "@/lib/api";
import { getAccessToken } from "@/lib/auth-storage";

export default function NewPostPage() {
  const router = useRouter();
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [ok, setOk] = useState(false);
  const [loading, setLoading] = useState(false);
  const [hasToken, setHasToken] = useState<boolean | null>(null);

  useEffect(() => {
    setHasToken(!!getAccessToken());
  }, []);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setOk(false);
    const token = getAccessToken();
    if (!token) {
      setError("请先登录");
      return;
    }
    setLoading(true);
    try {
      const body = await apiCreatePost(token, { title, content });
      if (body.code !== API_SUCCESS_CODE) {
        setError(apiErrorMessage(body));
        return;
      }
      setOk(true);
      setTitle("");
      setContent("");
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

      <form onSubmit={onSubmit} className="flex flex-col gap-4">
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
          disabled={loading}
          className="rounded-lg bg-zinc-900 py-2.5 text-sm font-medium text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900"
        >
          {loading ? "提交中…" : "发布"}
        </button>
      </form>
    </div>
  );
}
