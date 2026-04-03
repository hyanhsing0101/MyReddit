"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import {
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiLogin,
} from "@/lib/api";
import { setTokens } from "@/lib/auth-storage";

export default function LoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const body = await apiLogin({ username, password });
      if (body.code !== API_SUCCESS_CODE || !body.data) {
        setError(apiErrorMessage(body));
        return;
      }
      setTokens(body.data.access_token, body.data.refresh_token);
      router.push("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "请求失败");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="mx-auto flex min-h-full max-w-md flex-col justify-center px-4 py-16">
      <h1 className="mb-6 text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
        登录
      </h1>
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-zinc-600 dark:text-zinc-400">用户名</span>
          <input
            className="rounded-lg border border-zinc-300 bg-white px-3 py-2 dark:border-zinc-600 dark:bg-zinc-900"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            autoComplete="username"
            required
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-zinc-600 dark:text-zinc-400">密码</span>
          <input
            type="password"
            className="rounded-lg border border-zinc-300 bg-white px-3 py-2 dark:border-zinc-600 dark:bg-zinc-900"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
            required
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
          {loading ? "登录中…" : "登录"}
        </button>
      </form>
      <p className="mt-6 text-center text-sm text-zinc-500">
        没有账号？{" "}
        <Link href="/signup" className="text-zinc-900 underline dark:text-zinc-100">
          注册
        </Link>
      </p>
      <p className="mt-4 text-center">
        <Link href="/" className="text-sm text-zinc-500 underline">
          返回首页
        </Link>
      </p>
    </div>
  );
}
