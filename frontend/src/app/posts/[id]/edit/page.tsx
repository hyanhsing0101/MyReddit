"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { MarkdownEditor } from "@/components/markdown-editor";
import {
  API_FORBIDDEN_CODE,
  API_POST_NOT_EXIST_CODE,
  API_SUCCESS_CODE,
  API_TAG_COUNT_EXCEEDED_CODE,
  API_TAG_NOT_EXIST_CODE,
  apiErrorMessage,
  apiGetPost,
  apiListTags,
  apiMePermissions,
  apiUpdatePost,
  tagDisplayLabel,
  type MePermissionsPayload,
  type PostItem,
  type TagItem,
} from "@/lib/api";
import { getAccessToken } from "@/lib/auth-storage";

function canEditPost(post: PostItem, me: MePermissionsPayload | null): boolean {
  if (!me) return false;
  if (me.is_site_admin) return true;
  if (post.author_id == null) return false;
  return post.author_id === me.user_id;
}

export default function EditPostPage() {
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
  const [tags, setTags] = useState<TagItem[]>([]);
  const [tagsLoading, setTagsLoading] = useState(true);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [ok, setOk] = useState(false);
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([]);
  const [me, setMe] = useState<MePermissionsPayload | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setTagsLoading(true);
      try {
        const body = await apiListTags(1, 100);
        if (cancelled) return;
        if (body.code === API_SUCCESS_CODE && body.data) {
          setTags(body.data.list);
        } else {
          setError(apiErrorMessage(body));
        }
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : "加载标签失败");
      } finally {
        if (!cancelled) setTagsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!Number.isFinite(id) || id < 1) {
      setLoading(false);
      setError("无效的帖子 ID");
      return;
    }
    const token = getAccessToken();
    if (!token) {
      setLoading(false);
      setError("请先登录");
      return;
    }
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      try {
        const [postBody, meBody] = await Promise.all([
          apiGetPost(id, token),
          apiMePermissions(token),
        ]);
        if (cancelled) return;
        if (postBody.code === API_POST_NOT_EXIST_CODE || !postBody.data) {
          setError("帖子不存在或已删除");
          return;
        }
        if (postBody.code !== API_SUCCESS_CODE) {
          setError(apiErrorMessage(postBody));
          return;
        }
        const p = postBody.data;
        setPost(p);
        setTitle(p.title);
        setContent(p.content);
        setSelectedTagIds(Array.isArray(p.tags) ? p.tags.map((t) => t.id) : []);
        if (meBody.code === API_SUCCESS_CODE && meBody.data) {
          setMe(meBody.data);
        } else {
          setMe(null);
        }
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : "加载失败");
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [id]);

  const editable = useMemo(() => {
    if (!post) return false;
    return canEditPost(post, me);
  }, [post, me]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setOk(false);
    const token = getAccessToken();
    if (!token || !post) {
      setError("请先登录");
      return;
    }
    setSaving(true);
    try {
      const body = await apiUpdatePost(token, post.id, {
        title,
        content,
        tag_ids: selectedTagIds,
      });
      if (body.code === API_POST_NOT_EXIST_CODE) {
        setError("帖子不存在或已删除");
        return;
      }
      if (body.code === API_FORBIDDEN_CODE) {
        setError("无权限编辑该帖");
        return;
      }
      if (body.code === API_TAG_NOT_EXIST_CODE) {
        setError("包含无效标签，请刷新标签后重试");
        return;
      }
      if (body.code === API_TAG_COUNT_EXCEEDED_CODE) {
        setError("标签数量超限，最多 5 个");
        return;
      }
      if (body.code !== API_SUCCESS_CODE) {
        setError(apiErrorMessage(body));
        return;
      }
      setOk(true);
      setTimeout(() => {
        router.push(`/posts/${post.id}`);
      }, 400);
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存失败");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return <div className="mx-auto max-w-2xl px-4 py-16 text-sm text-zinc-500">加载中…</div>;
  }

  if (error && !post) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-16">
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        <p className="mt-4">
          <Link href="/" className="text-sm underline">返回首页</Link>
        </p>
      </div>
    );
  }

  if (!post || !editable) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-16">
        <p className="text-sm text-zinc-700 dark:text-zinc-300">无权限编辑该帖子。</p>
        <p className="mt-4">
          <Link href={post ? `/posts/${post.id}` : "/"} className="text-sm underline">
            返回
          </Link>
        </p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-12">
      <div className="mb-8 flex items-center justify-between gap-4">
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">编辑帖子</h1>
        <Link href={`/posts/${post.id}`} className="text-sm text-zinc-500 underline">返回帖子</Link>
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
        <div className="flex flex-col gap-2 text-sm">
          <span className="text-zinc-600 dark:text-zinc-400">正文（Markdown）</span>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            展示端经白名单过滤；图片可本地上传或 https 外链。
          </p>
          <MarkdownEditor
            value={content}
            onChange={setContent}
            accessToken={getAccessToken()}
            disabled={saving}
            rows={16}
          />
        </div>
        <fieldset className="flex flex-col gap-2 text-sm">
          <legend className="text-zinc-600 dark:text-zinc-400">标签（最多 5 个）</legend>
          {tagsLoading ? (
            <p className="text-xs text-zinc-500">加载标签中…</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {tags.map((tag) => {
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
                          setSelectedTagIds((prev) => prev.filter((id) => id !== tag.id));
                        }
                      }}
                    />
                    <span>#{tagDisplayLabel(tag)}</span>
                  </label>
                );
              })}
            </div>
          )}
        </fieldset>
        {error ? <p className="text-sm text-red-600 dark:text-red-400">{error}</p> : null}
        {ok ? <p className="text-sm text-emerald-700 dark:text-emerald-400">保存成功，正在返回…</p> : null}
        <button
          type="submit"
          disabled={saving}
          className="rounded-lg bg-zinc-900 py-2.5 text-sm font-medium text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900"
        >
          {saving ? "保存中…" : "保存修改"}
        </button>
      </form>
    </div>
  );
}
