"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { API_SUCCESS_CODE, apiCreateBoard, apiErrorMessage } from "@/lib/api";
import { getAccessToken } from "@/lib/auth-storage";

export default function NewBoardPage() {
  const router = useRouter();
  const [slug, setSlug] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [visibility, setVisibility] = useState<"public" | "private">("public");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [hasToken, setHasToken] = useState<boolean | null>(null);

  useEffect(() => {
    setHasToken(!!getAccessToken());
  }, []);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    const token = getAccessToken();
    if (!token) {
      setError("请先登录");
      return;
    }
    setLoading(true);
    try {
      const body = await apiCreateBoard(token, {
        slug,
        name,
        description: description.trim() || undefined,
        visibility,
      });
      if (body.code !== API_SUCCESS_CODE) {
        setError(apiErrorMessage(body));
        return;
      }
      router.push(`/boards/${encodeURIComponent(slug.trim().toLowerCase())}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建失败");
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
        <p className="text-zinc-700 dark:text-zinc-300">创建板块需要先登录。</p>
        <Link
          href="/login"
          className="mt-4 inline-block text-sm text-zinc-900 underline dark:text-zinc-100"
        >
          去登录
        </Link>
        <p className="mt-6">
          <Link href="/boards" className="text-sm text-zinc-500 underline">
            返回板块列表
          </Link>
        </p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-16">
      <div className="mb-8 flex items-center justify-between gap-4">
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
          创建板块
        </h1>
        <Link href="/boards" className="text-sm text-zinc-500 underline">
          返回列表
        </Link>
      </div>

      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-zinc-600 dark:text-zinc-400">
            Slug（URL 用，小写字母/数字/下划线，首字符须为字母或数字）
          </span>
          <input
            className="rounded-lg border border-zinc-300 bg-white px-3 py-2 font-mono text-sm dark:border-zinc-600 dark:bg-zinc-900"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            required
            maxLength={64}
            placeholder="例如 my_topic"
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-zinc-600 dark:text-zinc-400">展示名称</span>
          <input
            className="rounded-lg border border-zinc-300 bg-white px-3 py-2 dark:border-zinc-600 dark:bg-zinc-900"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            maxLength={128}
          />
        </label>
        <fieldset className="flex flex-col gap-2 text-sm">
          <legend className="text-zinc-600 dark:text-zinc-400">可见性</legend>
          <label className="flex cursor-pointer items-center gap-2">
            <input
              type="radio"
              name="visibility"
              checked={visibility === "public"}
              onChange={() => setVisibility("public")}
            />
            <span>公开：任何人可读；登录即可发帖；不可订阅</span>
          </label>
          <label className="flex cursor-pointer items-center gap-2">
            <input
              type="radio"
              name="visibility"
              checked={visibility === "private"}
              onChange={() => setVisibility("private")}
            />
            <span>私有：仅订阅成员与版主可读帖、发帖</span>
          </label>
        </fieldset>
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-zinc-600 dark:text-zinc-400">描述（可选）</span>
          <textarea
            className="min-h-[120px] rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm dark:border-zinc-600 dark:bg-zinc-900"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={4}
          />
        </label>
        {error ? (
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        ) : null}
        <button
          type="submit"
          disabled={loading}
          className="rounded-lg bg-zinc-900 py-2.5 text-sm font-medium text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900"
        >
          {loading ? "创建中…" : "创建"}
        </button>
      </form>
    </div>
  );
}
