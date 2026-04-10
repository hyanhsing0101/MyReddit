"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import {
  API_BOARD_NOT_EXIST_CODE,
  API_BOARD_MODERATOR_NOT_EXIST_CODE,
  API_CANNOT_REMOVE_LAST_OWNER_CODE,
  API_FORBIDDEN_CODE,
  API_SUCCESS_CODE,
  API_USER_NOT_EXIST_CODE,
  apiAddBoardModerator,
  apiErrorMessage,
  apiGetBoardBySlug,
  apiListBoardModerators,
  apiRemoveBoardModerator,
  apiUpdateBoardModeratorRole,
  type BoardItem,
  type BoardModeratorItem,
  type BoardModeratorRole,
} from "@/lib/api";
import { getAccessToken } from "@/lib/auth-storage";

export default function BoardModeratorsPage() {
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
  const [rows, setRows] = useState<BoardModeratorItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const [newUserID, setNewUserID] = useState("");
  const [newRole, setNewRole] = useState<BoardModeratorRole>("moderator");

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
      if (b.code !== API_SUCCESS_CODE || !b.data) {
        setError(apiErrorMessage(b));
        setBoard(null);
        setRows([]);
        return;
      }
      setBoard(b.data);

      const list = await apiListBoardModerators(token, b.data.id);
      if (list.code !== API_SUCCESS_CODE || !list.data) {
        setError(apiErrorMessage(list));
        setRows([]);
        return;
      }
      setRows(Array.isArray(list.data.list) ? list.data.list : []);
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载失败");
      setRows([]);
    } finally {
      setLoading(false);
    }
  }, [slug]);

  useEffect(() => {
    void load();
  }, [load]);

  async function onAdd(e: React.FormEvent) {
    e.preventDefault();
    if (!board) return;
    const token = getAccessToken();
    if (!token) {
      setError("请先登录");
      return;
    }
    const uid = Number.parseInt(newUserID, 10);
    if (!Number.isFinite(uid) || uid < 1) {
      setError("请输入合法 user_id");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const res = await apiAddBoardModerator(token, board.id, {
        user_id: uid,
        role: newRole,
      });
      if (res.code !== API_SUCCESS_CODE) {
        if (res.code === API_USER_NOT_EXIST_CODE) {
          setError("目标用户不存在");
          return;
        }
        if (res.code === API_BOARD_NOT_EXIST_CODE) {
          setError("板块不存在或已删除");
          return;
        }
        if (res.code === API_FORBIDDEN_CODE) {
          setError("无权限管理版主");
          return;
        }
        setError(apiErrorMessage(res));
        return;
      }
      setNewUserID("");
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "操作失败");
    } finally {
      setBusy(false);
    }
  }

  async function onChangeRole(userID: number, role: BoardModeratorRole) {
    if (!board) return;
    const token = getAccessToken();
    if (!token) {
      setError("请先登录");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const res = await apiUpdateBoardModeratorRole(token, board.id, userID, role);
      if (res.code !== API_SUCCESS_CODE) {
        if (res.code === API_BOARD_MODERATOR_NOT_EXIST_CODE) {
          setError("目标版主不存在，列表可能已变化");
          await load();
          return;
        }
        if (res.code === API_CANNOT_REMOVE_LAST_OWNER_CODE) {
          setError("不能移除最后一个 owner");
          return;
        }
        if (res.code === API_BOARD_NOT_EXIST_CODE) {
          setError("板块不存在或已删除");
          return;
        }
        if (res.code === API_FORBIDDEN_CODE) {
          setError("无权限管理版主");
          return;
        }
        setError(apiErrorMessage(res));
        return;
      }
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "操作失败");
    } finally {
      setBusy(false);
    }
  }

  async function onRemove(userID: number) {
    if (!board) return;
    const token = getAccessToken();
    if (!token) {
      setError("请先登录");
      return;
    }
    if (!window.confirm(`确认移除用户 ${userID} 的版主身份？`)) return;
    setBusy(true);
    setError(null);
    try {
      const res = await apiRemoveBoardModerator(token, board.id, userID);
      if (res.code !== API_SUCCESS_CODE) {
        if (res.code === API_BOARD_MODERATOR_NOT_EXIST_CODE) {
          setError("目标版主不存在，列表可能已变化");
          await load();
          return;
        }
        if (res.code === API_CANNOT_REMOVE_LAST_OWNER_CODE) {
          setError("不能移除最后一个 owner");
          return;
        }
        if (res.code === API_FORBIDDEN_CODE) {
          setError("无权限操作");
        } else {
          setError(apiErrorMessage(res));
        }
        return;
      }
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "操作失败");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-10">
      <button
        type="button"
        onClick={() => router.push(`/boards/${encodeURIComponent(slug)}`)}
        className="mb-6 text-sm text-zinc-500 underline"
      >
        ← 返回板块详情
      </button>

      <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
        版主管理
      </h1>
      {board ? (
        <p className="mt-1 text-sm text-zinc-500">
          板块：{board.name} /{board.slug}
        </p>
      ) : null}

      {loading ? <p className="mt-6 text-sm text-zinc-500">加载中…</p> : null}
      {error ? (
        <p className="mt-4 text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : null}

      {!loading && board ? (
        <>
          <form onSubmit={onAdd} className="mt-6 rounded-xl border border-zinc-200 p-4 dark:border-zinc-800">
            <h2 className="text-sm font-medium text-zinc-900 dark:text-zinc-100">新增/更新版主</h2>
            <div className="mt-3 flex flex-wrap items-end gap-3">
              <label className="flex flex-col text-sm">
                <span className="mb-1 text-zinc-600 dark:text-zinc-400">用户 ID</span>
                <input
                  className="rounded border border-zinc-300 px-3 py-2 dark:border-zinc-600 dark:bg-zinc-900"
                  value={newUserID}
                  onChange={(e) => setNewUserID(e.target.value)}
                  placeholder="如 10002"
                />
              </label>
              <label className="flex flex-col text-sm">
                <span className="mb-1 text-zinc-600 dark:text-zinc-400">角色</span>
                <select
                  className="rounded border border-zinc-300 px-3 py-2 dark:border-zinc-600 dark:bg-zinc-900"
                  value={newRole}
                  onChange={(e) => setNewRole(e.target.value as BoardModeratorRole)}
                >
                  <option value="moderator">moderator</option>
                  <option value="owner">owner</option>
                </select>
              </label>
              <button
                type="submit"
                disabled={busy}
                className="rounded bg-zinc-900 px-4 py-2 text-sm text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900"
              >
                保存
              </button>
            </div>
          </form>

          <section className="mt-6 rounded-xl border border-zinc-200 dark:border-zinc-800">
            <div className="border-b border-zinc-200 px-4 py-3 text-sm font-medium dark:border-zinc-800">
              版主列表
            </div>
            <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
              {rows.length === 0 ? (
                <p className="px-4 py-6 text-sm text-zinc-500">暂无版主</p>
              ) : (
                rows.map((r) => (
                  <div key={r.user_id} className="flex flex-wrap items-center justify-between gap-3 px-4 py-3">
                    <div className="text-sm">
                      <p className="font-medium text-zinc-900 dark:text-zinc-100">
                        {r.username || `用户 ${r.user_id}`}（{r.user_id}）
                      </p>
                      <p className="text-zinc-500">
                        角色：{r.role} · 任命人：{r.appointed_by ?? "-"}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        disabled={busy || r.role === "moderator"}
                        onClick={() => void onChangeRole(r.user_id, "moderator")}
                        className="rounded border border-zinc-300 px-3 py-1 text-xs disabled:opacity-40 dark:border-zinc-600"
                      >
                        设为 moderator
                      </button>
                      <button
                        type="button"
                        disabled={busy || r.role === "owner"}
                        onClick={() => void onChangeRole(r.user_id, "owner")}
                        className="rounded border border-zinc-300 px-3 py-1 text-xs disabled:opacity-40 dark:border-zinc-600"
                      >
                        设为 owner
                      </button>
                      <button
                        type="button"
                        disabled={busy}
                        onClick={() => void onRemove(r.user_id)}
                        className="rounded border border-red-300 px-3 py-1 text-xs text-red-700 disabled:opacity-40 dark:border-red-800 dark:text-red-300"
                      >
                        移除
                      </button>
                    </div>
                  </div>
                ))
              )}
            </div>
          </section>

          <p className="mt-6 text-sm text-zinc-500">
            说明：需要 owner 或站主权限才能改动；最后一个 owner 不能移除或降级。
          </p>
          <p className="mt-2">
            <Link href="/boards" className="text-sm text-zinc-500 underline">
              板块列表
            </Link>
          </p>
        </>
      ) : null}
    </div>
  );
}
