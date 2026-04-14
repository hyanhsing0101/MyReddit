const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://127.0.0.1:8081";

export const API_SUCCESS_CODE = 1000;
/** 与后端 controller.CodeNeedLogin 一致 */
export const API_NEED_LOGIN_CODE = 1007;
/** 与后端 controller.CodePostNotExist 一致 */
export const API_POST_NOT_EXIST_CODE = 1008;
/** 与后端 controller.CodeBoardNotExist 一致 */
export const API_BOARD_NOT_EXIST_CODE = 1009;
/** 与后端 controller.CodeBoardSlugTaken 一致 */
export const API_BOARD_SLUG_TAKEN_CODE = 1010;
/** 与后端 controller.CodeUserNotExist 一致 */
export const API_USER_NOT_EXIST_CODE = 1003;
/** 与后端 controller.CodeForbidden 一致 */
export const API_FORBIDDEN_CODE = 1011;
/** 与后端 controller.CodeNotBoardMember 一致 */
export const API_NOT_BOARD_MEMBER_CODE = 1012;
/** 与后端 controller.CodeCannotFavoritePublicBoard 一致 */
export const API_CANNOT_FAVORITE_PUBLIC_BOARD_CODE = 1013;
/** 与后端 controller.CodePostSealed 一致 */
export const API_POST_SEALED_CODE = 1014;
/** 与后端 controller.CodeCommentNotExist 一致 */
export const API_COMMENT_NOT_EXIST_CODE = 1015;
/** 与后端 controller.CodeInvalidVoteValue 一致 */
export const API_INVALID_VOTE_VALUE_CODE = 1016;
/** 与后端 controller.CodeTagNotExist 一致 */
export const API_TAG_NOT_EXIST_CODE = 1017;
/** 与后端 controller.CodeTagCountExceeded 一致 */
export const API_TAG_COUNT_EXCEEDED_CODE = 1018;
/** 与后端 controller.CodeCannotPostToSystemBoard 一致 */
export const API_CANNOT_POST_TO_SYSTEM_BOARD_CODE = 1019;
/** 与后端 controller.CodeBoardModeratorNotExist 一致 */
export const API_BOARD_MODERATOR_NOT_EXIST_CODE = 1020;
/** 与后端 controller.CodeCannotRemoveLastOwner 一致 */
export const API_CANNOT_REMOVE_LAST_OWNER_CODE = 1021;
/** 与后端 controller.CodeInvalidCommentParent 一致 */
export const API_INVALID_COMMENT_PARENT_CODE = 1022;
/** 与后端 controller.CodeParentCommentMismatch 一致 */
export const API_PARENT_COMMENT_MISMATCH_CODE = 1023;
/** 与后端 controller.CodeInvalidBoardID 一致 */
export const API_INVALID_BOARD_ID_CODE = 1024;
/** 与后端 controller.CodePostCommentsLocked 一致 */
export const API_POST_COMMENTS_LOCKED_CODE = 1025;
/** 与后端 controller.CodeCannotReportOwnPost 一致 */
export const API_CANNOT_REPORT_OWN_POST_CODE = 1026;
/** 与后端 controller.CodeDuplicatePostReport 一致 */
export const API_DUPLICATE_POST_REPORT_CODE = 1027;
/** 与后端 controller.CodePostReportNotExist 一致 */
export const API_POST_REPORT_NOT_EXIST_CODE = 1028;
/** 与后端 controller.CodeCannotAppealUnsealedPost 一致 */
export const API_CANNOT_APPEAL_UNSEALED_POST_CODE = 1029;
/** 与后端 controller.CodePostAppealNotExist 一致 */
export const API_POST_APPEAL_NOT_EXIST_CODE = 1030;
/** 与后端 controller.CodePostNotSoftDeleted 一致 */
export const API_POST_NOT_SOFT_DELETED_CODE = 1031;
/** 与后端 controller.CodeCannotReportOwnComment 一致 */
export const API_CANNOT_REPORT_OWN_COMMENT_CODE = 1032;
/** 与后端 controller.CodeDuplicateCommentReport 一致 */
export const API_DUPLICATE_COMMENT_REPORT_CODE = 1033;
/** 与后端 controller.CodeCommentReportNotExist 一致 */
export const API_COMMENT_REPORT_NOT_EXIST_CODE = 1034;

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

export type PostModerationActions = {
  can_seal: boolean;
  can_unseal: boolean;
  can_lock_comments: boolean;
  can_unlock_comments: boolean;
  can_pin: boolean;
  can_unpin: boolean;
};

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
  /** 带合法 Bearer 时由后端返回 */
  is_favorited?: boolean;
  sealed?: boolean;
  seal_kind?: string;
  comments_locked?: boolean;
  pinned?: boolean;
  moderation_actions?: PostModerationActions;
  create_time: string;
  update_time: string;
};

export type PostListPayload = {
  list: PostItem[];
  total: number;
  page: number;
  page_size: number;
};

/** 与后端 GET /posts?sort= 一致 */
export type PostSort = "new" | "hot" | "top";

/** 全站列表或仅已收藏板块（订阅流，须登录）。 */
export type PostListFeed = "all" | "subscribed";

export async function apiListPosts(
  page = 1,
  pageSize = 10,
  boardId?: number,
  accessToken?: string | null,
  sort: PostSort = "new",
  feed: PostListFeed = "all",
): Promise<ApiResponse<PostListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
    sort,
  });
  if (feed === "subscribed") {
    q.set("feed", "subscribed");
  }
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
  score: number;
  my_vote?: number | null;
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
  accessToken?: string | null,
): Promise<ApiResponse<CommentListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  const headers: HeadersInit = {};
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }
  const res = await fetch(`${API_BASE}/posts/${postId}/comments?${q.toString()}`, {
    headers,
  });
  return parseJson<CommentListPayload>(res);
}

export async function apiVoteComment(
  accessToken: string,
  postId: number,
  commentId: number,
  value: 1 | -1 | 0,
): Promise<ApiResponse<PostVotePayload>> {
  const res = await fetch(
    `${API_BASE}/posts/${postId}/comments/${commentId}/vote`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify({ value }),
    },
  );
  return parseJson<PostVotePayload>(res);
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

export async function apiSealPost(
  accessToken: string,
  postId: number,
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/seal`, {
    method: "POST",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<null>(res);
}

export async function apiUnsealPost(
  accessToken: string,
  postId: number,
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/seal`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<null>(res);
}

export async function apiLockPostComments(
  accessToken: string,
  postId: number,
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/lock-comments`, {
    method: "POST",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<null>(res);
}

export async function apiUnlockPostComments(
  accessToken: string,
  postId: number,
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/lock-comments`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<null>(res);
}

export async function apiPinPost(
  accessToken: string,
  postId: number,
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/pin`, {
    method: "POST",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<null>(res);
}

export async function apiUnpinPost(
  accessToken: string,
  postId: number,
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/pin`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<null>(res);
}

export async function apiCreatePostReport(
  accessToken: string,
  postId: number,
  payload: { reason: string; detail?: string },
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/reports`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export async function apiCreateCommentReport(
  accessToken: string,
  postId: number,
  commentId: number,
  payload: { reason: string; detail?: string },
): Promise<ApiResponse<null>> {
  const res = await fetch(
    `${API_BASE}/posts/${postId}/comments/${commentId}/reports`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(payload),
    },
  );
  return parseJson<null>(res);
}

export async function apiGetMyPostAppeal(
  accessToken: string,
  postId: number,
): Promise<ApiResponse<PostAppealItem>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/appeals/me`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<PostAppealItem>(res);
}

export async function apiUpsertMyPostAppeal(
  accessToken: string,
  postId: number,
  payload: {
    reason: string;
    requested_title: string;
    requested_content: string;
    user_reply?: string;
  },
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/appeals/me`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export async function apiAddPostFavorite(
  accessToken: string,
  postId: number,
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/favorites`, {
    method: "POST",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<null>(res);
}

export async function apiRemovePostFavorite(
  accessToken: string,
  postId: number,
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/favorites`, {
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
  moderated_board_ids: number[];
};

export type UserHomePostItem = {
  id: number;
  board_id: number;
  board_slug: string;
  board_name: string;
  title: string;
  score: number;
  create_time: string;
  update_time: string;
};

export type UserHomeCommentItem = {
  id: number;
  post_id: number;
  post_title: string;
  content: string;
  score: number;
  create_time: string;
  update_time: string;
};

export type UserHomePayload = {
  user_id: number;
  username: string;
  posts: UserHomePostItem[];
  posts_total: number;
  post_page: number;
  post_page_size: number;
  comments: UserHomeCommentItem[];
  comments_total: number;
  comment_page: number;
  comment_page_size: number;
};

export async function apiGetUserHome(
  userId: number,
  postPage = 1,
  postPageSize = 10,
  commentPage = 1,
  commentPageSize = 10,
  accessToken?: string | null,
): Promise<ApiResponse<UserHomePayload>> {
  const q = new URLSearchParams({
    post_page: String(postPage),
    post_page_size: String(postPageSize),
    comment_page: String(commentPage),
    comment_page_size: String(commentPageSize),
  });
  const headers: HeadersInit = {};
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }
  const res = await fetch(`${API_BASE}/users/${userId}/home?${q.toString()}`, {
    headers,
  });
  return parseJson<UserHomePayload>(res);
}

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
  const res = await fetch(`${API_BASE}/posts`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export async function apiUpdatePost(
  accessToken: string,
  postId: number,
  payload: { tag_ids: number[]; title: string; content: string },
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}`, {
    method: "PATCH",
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
  /** public | private */
  visibility: string;
  is_system_sink: boolean;
  create_time: string;
  update_time: string;
  /** 带合法 Bearer 时由后端返回 */
  is_favorited?: boolean;
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
  accessToken?: string | null,
): Promise<ApiResponse<BoardListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  if (includeSystemSink) q.set("include_system_sink", "true");
  const headers: HeadersInit = {};
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }
  const res = await fetch(`${API_BASE}/boards?${q.toString()}`, { headers });
  return parseJson<BoardListPayload>(res);
}

export async function apiGetBoardBySlug(
  slug: string,
  accessToken?: string | null,
): Promise<ApiResponse<BoardItem>> {
  const headers: HeadersInit = {};
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }
  const res = await fetch(
    `${API_BASE}/boards/by-slug/${encodeURIComponent(slug)}`,
    { headers },
  );
  return parseJson<BoardItem>(res);
}

export type BoardFavoriteRow = BoardItem & { favorited_at: string };

export type BoardModeratorRole = "owner" | "moderator";

export type BoardModeratorItem = {
  user_id: number;
  username: string;
  role: BoardModeratorRole;
  appointed_by: number | null;
  create_time: string;
  update_time: string;
};

export type BoardModeratorListPayload = {
  list: BoardModeratorItem[];
};

export type BoardFavoriteListPayload = {
  list: BoardFavoriteRow[];
  total: number;
  page: number;
  page_size: number;
};

export type PostFavoriteRow = PostItem & { favorited_at: string };

export type PostFavoriteListPayload = {
  list: PostFavoriteRow[];
  total: number;
  page: number;
  page_size: number;
};

export type PostReportStatus = "open" | "in_review" | "resolved" | "rejected";

export type PostReportItem = {
  id: number;
  post_id: number;
  post_title: string;
  board_id: number;
  reporter_id: number;
  reporter_username: string;
  reason: string;
  detail: string;
  status: PostReportStatus;
  handler_id: number | null;
  handler_username: string;
  handler_note: string;
  create_time: string;
  update_time: string;
};

export type PostReportListPayload = {
  list: PostReportItem[];
  total: number;
  pending_open: number;
  page: number;
  page_size: number;
};

/** 与帖子举报工单状态字符串一致 */
export type CommentReportStatus = PostReportStatus;

export type CommentReportItem = {
  id: number;
  comment_id: number;
  post_id: number;
  post_title: string;
  comment_snippet: string;
  board_id: number;
  reporter_id: number;
  reporter_username: string;
  reason: string;
  detail: string;
  status: CommentReportStatus;
  handler_id: number | null;
  handler_username: string;
  handler_note: string;
  create_time: string;
  update_time: string;
};

export type CommentReportListPayload = {
  list: CommentReportItem[];
  total: number;
  pending_open: number;
  page: number;
  page_size: number;
};

export type DeletedPostRow = {
  id: number;
  title: string;
  author_id?: number;
  deleted_at: string;
};

export type DeletedPostListPayload = {
  list: DeletedPostRow[];
  total: number;
  page: number;
  page_size: number;
};

export type PostAppealStatus = "open" | "in_review" | "approved" | "rejected";

export type PostAppealItem = {
  id: number;
  post_id: number;
  post_title: string;
  board_id: number;
  author_id: number;
  author_username: string;
  reason: string;
  requested_title: string;
  requested_content: string;
  user_reply: string;
  status: PostAppealStatus;
  moderator_id: number | null;
  moderator_username: string;
  moderator_reply: string;
  create_time: string;
  update_time: string;
};

export type PostAppealListPayload = {
  list: PostAppealItem[];
  total: number;
  page: number;
  page_size: number;
};

export type ModerationDashboardPayload = {
  pending_reports: number;
  in_review_reports: number;
  resolved_reports_7d: number;
  rejected_reports_7d: number;
  logs_24h: number;
  reports_created_24h: number;
};

export type ModerationAction =
  | "seal_post"
  | "unseal_post"
  | "delete_post"
  | "restore_post"
  | "handle_post_appeal"
  | "lock_post_comments"
  | "unlock_post_comments"
  | "pin_post"
  | "unpin_post"
  | "update_post_report_status"
  | "update_comment_report_status"
  | "upsert_board_moderator"
  | "update_board_moderator_role"
  | "remove_board_moderator";

export type ModerationLogItem = {
  id: number;
  board_id: number;
  operator_id: number;
  operator_username: string;
  action: ModerationAction;
  target_type: "post" | "post_report" | "comment_report" | "board_moderator";
  target_id: number;
  description: string;
  create_time: string;
};

export type ModerationLogListPayload = {
  list: ModerationLogItem[];
  total: number;
  page: number;
  page_size: number;
};

export async function apiListFavoriteBoards(
  accessToken: string,
  page = 1,
  pageSize = 20,
): Promise<ApiResponse<BoardFavoriteListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  const res = await fetch(`${API_BASE}/me/favorite-boards?${q.toString()}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<BoardFavoriteListPayload>(res);
}

export async function apiListFavoritePosts(
  accessToken: string,
  page = 1,
  pageSize = 20,
): Promise<ApiResponse<PostFavoriteListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  const res = await fetch(`${API_BASE}/me/favorite-posts?${q.toString()}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<PostFavoriteListPayload>(res);
}

export async function apiListBoardReports(
  accessToken: string,
  boardId: number,
  page = 1,
  pageSize = 20,
  status?: PostReportStatus | "",
): Promise<ApiResponse<PostReportListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  if (status) q.set("status", status);
  const res = await fetch(`${API_BASE}/boards/${boardId}/reports?${q.toString()}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<PostReportListPayload>(res);
}

export async function apiBatchUpdateBoardReports(
  accessToken: string,
  boardId: number,
  payload: { report_ids: number[]; status: PostReportStatus; handler_note?: string },
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/boards/${boardId}/reports`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export async function apiUpdateBoardReportStatus(
  accessToken: string,
  boardId: number,
  reportId: number,
  payload: { status: PostReportStatus; handler_note?: string },
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/boards/${boardId}/reports/${reportId}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export async function apiListBoardCommentReports(
  accessToken: string,
  boardId: number,
  page = 1,
  pageSize = 20,
  status?: CommentReportStatus | "",
): Promise<ApiResponse<CommentReportListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  if (status) q.set("status", status);
  const res = await fetch(
    `${API_BASE}/boards/${boardId}/comment-reports?${q.toString()}`,
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
  return parseJson<CommentReportListPayload>(res);
}

export async function apiBatchUpdateBoardCommentReports(
  accessToken: string,
  boardId: number,
  payload: {
    report_ids: number[];
    status: CommentReportStatus;
    handler_note?: string;
  },
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/boards/${boardId}/comment-reports`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export async function apiUpdateBoardCommentReportStatus(
  accessToken: string,
  boardId: number,
  reportId: number,
  payload: { status: CommentReportStatus; handler_note?: string },
): Promise<ApiResponse<null>> {
  const res = await fetch(
    `${API_BASE}/boards/${boardId}/comment-reports/${reportId}`,
    {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(payload),
    },
  );
  return parseJson<null>(res);
}

export async function apiListBoardDeletedPosts(
  accessToken: string,
  boardId: number,
  page = 1,
  pageSize = 20,
): Promise<ApiResponse<DeletedPostListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  const res = await fetch(
    `${API_BASE}/boards/${boardId}/deleted-posts?${q.toString()}`,
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
  return parseJson<DeletedPostListPayload>(res);
}

export async function apiRestorePost(
  accessToken: string,
  postId: number,
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/posts/${postId}/restore`, {
    method: "POST",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<null>(res);
}

export async function apiListBoardAppeals(
  accessToken: string,
  boardId: number,
  page = 1,
  pageSize = 20,
  status?: PostAppealStatus | "",
): Promise<ApiResponse<PostAppealListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  if (status) q.set("status", status);
  const res = await fetch(`${API_BASE}/boards/${boardId}/appeals?${q.toString()}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<PostAppealListPayload>(res);
}

export async function apiHandleBoardAppeal(
  accessToken: string,
  boardId: number,
  appealId: number,
  payload: {
    status: "in_review" | "approved" | "rejected";
    moderator_reply?: string;
    apply_update?: boolean;
  },
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/boards/${boardId}/appeals/${appealId}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export async function apiListBoardModerationLogs(
  accessToken: string,
  boardId: number,
  page = 1,
  pageSize = 20,
  action?: ModerationAction | "",
  targetType?:
    | "post"
    | "post_report"
    | "comment_report"
    | "board_moderator"
    | "",
  targetID?: number | null,
): Promise<ApiResponse<ModerationLogListPayload>> {
  const q = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  if (action) q.set("action", action);
  if (targetType) q.set("target_type", targetType);
  if (targetID != null && targetID >= 1) q.set("target_id", String(targetID));
  const res = await fetch(`${API_BASE}/boards/${boardId}/mod-logs?${q.toString()}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<ModerationLogListPayload>(res);
}

export async function apiGetModerationDashboard(
  accessToken: string,
  boardId: number,
): Promise<ApiResponse<ModerationDashboardPayload>> {
  const res = await fetch(`${API_BASE}/boards/${boardId}/mod-dashboard`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<ModerationDashboardPayload>(res);
}

export async function apiAddBoardFavorite(
  accessToken: string,
  boardId: number,
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/boards/${boardId}/favorites`, {
    method: "POST",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<null>(res);
}

export async function apiRemoveBoardFavorite(
  accessToken: string,
  boardId: number,
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/boards/${boardId}/favorites`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<null>(res);
}

export async function apiCreateBoard(
  accessToken: string,
  payload: {
    slug: string;
    name: string;
    description?: string;
    visibility?: "public" | "private";
  },
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

export async function apiListBoardModerators(
  accessToken: string,
  boardId: number,
): Promise<ApiResponse<BoardModeratorListPayload>> {
  const res = await fetch(`${API_BASE}/boards/${boardId}/moderators`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return parseJson<BoardModeratorListPayload>(res);
}

export async function apiAddBoardModerator(
  accessToken: string,
  boardId: number,
  payload: { user_id: number; role: BoardModeratorRole },
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/boards/${boardId}/moderators`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export async function apiUpdateBoardModeratorRole(
  accessToken: string,
  boardId: number,
  userId: number,
  role: BoardModeratorRole,
): Promise<ApiResponse<null>> {
  const res = await fetch(
    `${API_BASE}/boards/${boardId}/moderators/${userId}`,
    {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify({ role }),
    },
  );
  return parseJson<null>(res);
}

export async function apiRemoveBoardModerator(
  accessToken: string,
  boardId: number,
  userId: number,
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/boards/${boardId}/moderators/${userId}`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${accessToken}` },
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
  accessToken?: string | null,
): Promise<ApiResponse<SearchDataPayload>> {
  const params = new URLSearchParams({
    q,
    scope,
    post_limit: String(postLimit),
    board_limit: String(boardLimit),
  });
  const headers: HeadersInit = {};
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }
  const res = await fetch(`${API_BASE}/search?${params.toString()}`, {
    headers,
  });
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
