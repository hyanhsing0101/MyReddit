const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://127.0.0.1:8081";

export const API_SUCCESS_CODE = 1000;
/** 与后端 controller.CodeNeedLogin 一致 */
export const API_NEED_LOGIN_CODE = 1007;
/** 与后端 controller.CodePostNotExist 一致 */
export const API_POST_NOT_EXIST_CODE = 1008;
/** 与后端 controller.CodeForbidden 一致 */
export const API_FORBIDDEN_CODE = 1011;

export type ApiResponse<T> = {
  code: number;
  msg: unknown;
  data: T;
};

async function parseJson<T>(res: Response): Promise<ApiResponse<T>> {
  const text = await res.text();
  try {
    return JSON.parse(text) as ApiResponse<T>;
  } catch {
    throw new Error(`无效响应 (${res.status}): ${text.slice(0, 200)}`);
  }
}

export function apiErrorMessage(body: ApiResponse<unknown>): string {
  const { msg } = body;
  if (typeof msg === "string") return msg;
  if (msg !== null && typeof msg === "object") {
    return JSON.stringify(msg);
  }
  return `错误码 ${body.code}`;
}

export async function apiSignup(payload: {
  username: string;
  password: string;
  re_password: string;
}): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/signup`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export async function apiLogin(payload: {
  username: string;
  password: string;
}): Promise<
  ApiResponse<{ access_token: string; refresh_token: string } | null>
> {
  const res = await fetch(`${API_BASE}/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return parseJson(res);
}

export type PostItem = {
  id: number;
  board_id: number;
  board_slug: string;
  board_name: string;
  tags: TagItem[];
  title: string;
  content: string;
  author_id: number | null;
  /** 净分（上票 − 下票） */
  score: number;
  /** 当前用户投票：1 / -1；未投为 null；未登录时通常不出现 */
  my_vote?: number | null;
  create_time: string;
  update_time: string;
};

export type PostListPayload = {
  list: PostItem[];
  total: number;
  page: number;
  page_size: number;
};

export async function apiListPosts(
  page = 1,
  pageSize = 10,
  boardId?: number,
  accessToken?: string | null,
): Promise<ApiResponse<PostListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  if (boardId != null && boardId >= 1) {
    q.set("board_id", String(boardId));
  }
  const headers: HeadersInit = {};
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }
  const res = await fetch(`${API_BASE}/posts?${q.toString()}`, { headers });
  return parseJson<PostListPayload>(res);
}

export async function apiGetPost(
  id: number,
  accessToken?: string | null,
): Promise<ApiResponse<PostItem>> {
  const headers: HeadersInit = {};
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }
  const res = await fetch(`${API_BASE}/posts/${id}`, { headers });
  return parseJson<PostItem>(res);
}

export type PostVotePayload = {
  score: number;
  my_vote: number | null;
};

export async function apiVotePost(
  accessToken: string,
  postId: number,
  value: 1 | -1 | 0,
): Promise<ApiResponse<PostVotePayload>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/vote`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify({ value }),
  });
  return parseJson<PostVotePayload>(res);
}

export type CommentItem = {
  id: number;
  post_id: number;
  author_id: number | null;
  author_username: string;
  parent_id: number | null;
  content: string;
  create_time: string;
  update_time: string;
};

export type CommentListPayload = {
  list: CommentItem[];
  total: number;
  page: number;
  page_size: number;
};

export async function apiListComments(
  postId: number,
  page = 1,
  pageSize = 100,
): Promise<ApiResponse<CommentListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  const res = await fetch(`${API_BASE}/posts/${postId}/comments?${q.toString()}`);
  return parseJson<CommentListPayload>(res);
}

export async function apiCreateComment(
  accessToken: string,
  postId: number,
  payload: { content: string; parent_id?: number },
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/comments`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export async function apiDeletePost(
  accessToken: string,
  postId: number,
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<null>(res);
}

export type MePermissionsPayload = {
  user_id: number;
  username: string;
  roles: string[];
  is_site_admin: boolean;
};

export async function apiMePermissions(
  accessToken: string,
): Promise<ApiResponse<MePermissionsPayload>> {
  const res = await fetch(`${API_BASE}/me/permissions`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<MePermissionsPayload>(res);
}

export async function apiCreatePost(
  accessToken: string,
  payload: { board_id: number; tag_ids: number[]; title: string; content: string },
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/post`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export type BoardItem = {
  id: number;
  slug: string;
  name: string;
  description: string;
  created_by: number | null;
  is_system_sink: boolean;
  create_time: string;
  update_time: string;
};

export type TagItem = {
  id: number;
  slug: string;
  name: string;
  description: string;
  create_time: string;
  update_time: string;
};

/** 展示用：优先 name，空则 slug */
export function tagDisplayLabel(tag: Pick<TagItem, "name" | "slug">): string {
  const n = tag.name?.trim();
  return n || tag.slug;
}

export type TagListPayload = {
  list: TagItem[];
  total: number;
  page: number;
  page_size: number;
};

export type BoardListPayload = {
  list: BoardItem[];
  total: number;
  page: number;
  page_size: number;
};

export type SearchScope = "all" | "posts" | "boards";

export type SearchDataPayload = {
  query: string;
  scope: SearchScope;
  posts: PostItem[];
  boards: BoardItem[];
};

export async function apiListBoards(
  page = 1,
  pageSize = 20,
  includeSystemSink = false,
): Promise<ApiResponse<BoardListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  if (includeSystemSink) q.set("include_system_sink", "true");
  const res = await fetch(`${API_BASE}/boards?${q.toString()}`);
  return parseJson<BoardListPayload>(res);
}

export async function apiGetBoardBySlug(
  slug: string,
): Promise<ApiResponse<BoardItem>> {
  const res = await fetch(
    `${API_BASE}/boards/slug/${encodeURIComponent(slug)}`,
  );
  return parseJson<BoardItem>(res);
}

export async function apiCreateBoard(
  accessToken: string,
  payload: { slug: string; name: string; description?: string },
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/boards`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export async function apiListTags(
  page = 1,
  pageSize = 50,
): Promise<ApiResponse<TagListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  const res = await fetch(`${API_BASE}/tags?${q.toString()}`);
  return parseJson<TagListPayload>(res);
}

export async function apiSearch(
  q: string,
  scope: SearchScope = "all",
  postLimit = 20,
  boardLimit = 10,
): Promise<ApiResponse<SearchDataPayload>> {
  const params = new URLSearchParams({
    q,
    scope,
    post_limit: String(postLimit),
    board_limit: String(boardLimit),
  });
  const res = await fetch(`${API_BASE}/search?${params.toString()}`);
  return parseJson<SearchDataPayload>(res);
}

export async function apiPing(accessToken: string): Promise<string> {
  const res = await fetch(`${API_BASE}/ping`, {
    method: "GET",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  const ct = res.headers.get("content-type") ?? "";
  if (ct.includes("application/json")) {
    const body = (await res.json()) as ApiResponse<unknown>;
    throw new Error(apiErrorMessage(body));
  }
  return res.text();
}
