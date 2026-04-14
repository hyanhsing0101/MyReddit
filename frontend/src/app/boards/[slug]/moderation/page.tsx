"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  API_BOARD_NOT_EXIST_CODE,
  API_COMMENT_REPORT_NOT_EXIST_CODE,
  API_FORBIDDEN_CODE,
  API_POST_APPEAL_NOT_EXIST_CODE,
  API_POST_NOT_SOFT_DELETED_CODE,
  API_POST_REPORT_NOT_EXIST_CODE,
  API_SUCCESS_CODE,
  apiBatchUpdateBoardCommentReports,
  apiErrorMessage,
  apiGetModerationDashboard,
  apiGetBoardBySlug,
  apiBatchUpdateBoardReports,
  apiHandleBoardAppeal,
  apiListBoardAppeals,
  apiListBoardCommentReports,
  apiListBoardDeletedPosts,
  apiListBoardModerationLogs,
  apiListBoardReports,
  apiRestorePost,
  apiUpdateBoardCommentReportStatus,
  apiUpdateBoardReportStatus,
  type BoardItem,
  type CommentReportItem,
  type CommentReportStatus,
  type DeletedPostRow,
  type ModerationDashboardPayload,
  type ModerationAction,
  type ModerationLogItem,
  type PostAppealItem,
  type PostAppealStatus,
  type PostReportItem,
  type PostReportStatus,
} from "@/lib/api";
import { getAccessToken } from "@/lib/auth-storage";

const REPORT_STATUS_OPTIONS: Array<{ id: "" | PostReportStatus; label: string }> = [
  { id: "", label: "全部" },
  { id: "open", label: "待处理" },
  { id: "in_review", label: "处理中" },
  { id: "resolved", label: "已解决" },
  { id: "rejected", label: "已驳回" },
];

const REPORT_STATUS_LABEL: Record<PostReportStatus, string> = {
  open: "待处理",
  in_review: "处理中",
  resolved: "已解决",
  rejected: "已驳回",
};

const COMMENT_REPORT_STATUS_OPTIONS: Array<{ id: "" | CommentReportStatus; label: string }> =
  REPORT_STATUS_OPTIONS;
const COMMENT_REPORT_STATUS_LABEL: Record<CommentReportStatus, string> =
  REPORT_STATUS_LABEL;

const APPEAL_STATUS_OPTIONS: Array<{ id: "" | PostAppealStatus; label: string }> = [
  { id: "", label: "全部" },
  { id: "open", label: "待处理" },
  { id: "in_review", label: "处理中" },
  { id: "approved", label: "已通过" },
  { id: "rejected", label: "已驳回" },
];

const APPEAL_STATUS_LABEL: Record<PostAppealStatus, string> = {
  open: "待处理",
  in_review: "处理中",
  approved: "已通过",
  rejected: "已驳回",
};

const LOG_ACTION_OPTIONS: Array<{ id: "" | ModerationAction; label: string }> = [
  { id: "", label: "全部动作" },
  { id: "update_post_report_status", label: "处理帖子举报" },
  { id: "update_comment_report_status", label: "处理评论举报" },
  { id: "seal_post", label: "封帖" },
  { id: "unseal_post", label: "解封" },
  { id: "delete_post", label: "软删帖子" },
  { id: "restore_post", label: "恢复帖子" },
  { id: "handle_post_appeal", label: "处理申诉" },
  { id: "lock_post_comments", label: "锁评" },
  { id: "unlock_post_comments", label: "解锁评论" },
  { id: "pin_post", label: "置顶" },
  { id: "unpin_post", label: "取消置顶" },
  { id: "upsert_board_moderator", label: "新增/更新版主" },
  { id: "update_board_moderator_role", label: "更新版主角色" },
  { id: "remove_board_moderator", label: "移除版主" },
];

const LOG_ACTION_LABEL: Record<ModerationAction, string> = {
  seal_post: "封帖",
  unseal_post: "解封",
  delete_post: "软删帖子",
  restore_post: "恢复帖子",
  handle_post_appeal: "处理申诉",
  lock_post_comments: "锁评",
  unlock_post_comments: "解锁评论",
  pin_post: "置顶",
  unpin_post: "取消置顶",
  update_post_report_status: "处理帖子举报",
  update_comment_report_status: "处理评论举报",
  upsert_board_moderator: "新增/更新版主",
  update_board_moderator_role: "更新版主角色",
  remove_board_moderator: "移除版主",
};

export default function BoardModerationPage() {
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
  const [loadingBoard, setLoadingBoard] = useState(true);
  const [pageError, setPageError] = useState<string | null>(null);

  const [reportRows, setReportRows] = useState<PostReportItem[]>([]);
  const [reportPage, setReportPage] = useState(1);
  const [reportTotal, setReportTotal] = useState(0);
  const [reportStatus, setReportStatus] = useState<"" | PostReportStatus>("open");
  const [pendingOpen, setPendingOpen] = useState(0);
  const [dashboard, setDashboard] = useState<ModerationDashboardPayload | null>(null);
  const [reportLoading, setReportLoading] = useState(false);
  const [reportBusyID, setReportBusyID] = useState<number | null>(null);
  const [focusReportID, setFocusReportID] = useState<number | null>(null);
  const [selectedReportIDs, setSelectedReportIDs] = useState<number[]>([]);
  const [batchStatus, setBatchStatus] = useState<PostReportStatus>("in_review");
  const [batchBusy, setBatchBusy] = useState(false);

  const [cReportRows, setCReportRows] = useState<CommentReportItem[]>([]);
  const [cReportPage, setCReportPage] = useState(1);
  const [cReportTotal, setCReportTotal] = useState(0);
  const [cReportStatus, setCReportStatus] = useState<"" | CommentReportStatus>("open");
  const [cPendingOpen, setCPendingOpen] = useState(0);
  const [cReportLoading, setCReportLoading] = useState(false);
  const [cReportBusyID, setCReportBusyID] = useState<number | null>(null);
  const [focusCommentReportID, setFocusCommentReportID] = useState<number | null>(
    null,
  );
  const [cSelectedReportIDs, setCSelectedReportIDs] = useState<number[]>([]);
  const [cBatchStatus, setCBatchStatus] = useState<CommentReportStatus>("in_review");
  const [cBatchBusy, setCBatchBusy] = useState(false);

  const [delRows, setDelRows] = useState<DeletedPostRow[]>([]);
  const [delPage, setDelPage] = useState(1);
  const [delTotal, setDelTotal] = useState(0);
  const [delLoading, setDelLoading] = useState(false);
  const [restoreBusyID, setRestoreBusyID] = useState<number | null>(null);

  const [appealRows, setAppealRows] = useState<PostAppealItem[]>([]);
  const [appealPage, setAppealPage] = useState(1);
  const [appealTotal, setAppealTotal] = useState(0);
  const [appealStatus, setAppealStatus] = useState<"" | PostAppealStatus>("open");
  const [appealLoading, setAppealLoading] = useState(false);
  const [appealBusyID, setAppealBusyID] = useState<number | null>(null);

  const [logRows, setLogRows] = useState<ModerationLogItem[]>([]);
  const [logPage, setLogPage] = useState(1);
  const [logTotal, setLogTotal] = useState(0);
  const [logAction, setLogAction] = useState<"" | ModerationAction>("");
  const [logTargetType, setLogTargetType] = useState<
    "" | "post" | "post_report" | "comment_report" | "board_moderator"
  >("");
  const [logTargetIDInput, setLogTargetIDInput] = useState("");
  const [logLoading, setLogLoading] = useState(false);

  const pageSize = 20;

  const loadBoard = useCallback(async () => {
    if (!slug) {
      setPageError("无效的板块地址");
      setLoadingBoard(false);
      return;
    }
    const token = getAccessToken();
    if (!token) {
      setPageError("请先登录");
      setLoadingBoard(false);
      return;
    }
    setLoadingBoard(true);
    setPageError(null);
    try {
      const b = await apiGetBoardBySlug(slug, token);
      if (b.code === API_BOARD_NOT_EXIST_CODE) {
        setPageError("板块不存在或无权访问");
        setBoard(null);
        return;
      }
      if (b.code !== API_SUCCESS_CODE || !b.data) {
        setPageError(apiErrorMessage(b));
        setBoard(null);
        return;
      }
      setBoard(b.data);
    } catch (e) {
      setPageError(e instanceof Error ? e.message : "加载失败");
      setBoard(null);
    } finally {
      setLoadingBoard(false);
    }
  }, [slug]);

  const loadReports = useCallback(async () => {
    if (!board) return;
    const token = getAccessToken();
    if (!token) return;
    setReportLoading(true);
    setPageError(null);
    try {
      const body = await apiListBoardReports(
        token,
        board.id,
        reportPage,
        pageSize,
        reportStatus,
      );
      if (body.code === API_FORBIDDEN_CODE) {
        setPageError("你不是该板块版主或站点管理员");
        setReportRows([]);
        setReportTotal(0);
        return;
      }
      if (body.code !== API_SUCCESS_CODE || !body.data) {
        setPageError(apiErrorMessage(body));
        setReportRows([]);
        setReportTotal(0);
        return;
      }
      setReportRows(Array.isArray(body.data.list) ? body.data.list : []);
      setReportTotal(body.data.total ?? 0);
      setPendingOpen(body.data.pending_open ?? 0);
      setReportPage(body.data.page ?? reportPage);
      setSelectedReportIDs((prev) =>
        prev.filter((id) => (body.data.list ?? []).some((r) => r.id === id)),
      );
    } catch (e) {
      setPageError(e instanceof Error ? e.message : "举报加载失败");
      setReportRows([]);
      setReportTotal(0);
    } finally {
      setReportLoading(false);
    }
  }, [board, reportPage, pageSize, reportStatus]);

  const loadCReports = useCallback(async () => {
    if (!board) return;
    const token = getAccessToken();
    if (!token) return;
    setCReportLoading(true);
    setPageError(null);
    try {
      const body = await apiListBoardCommentReports(
        token,
        board.id,
        cReportPage,
        pageSize,
        cReportStatus,
      );
      if (body.code === API_FORBIDDEN_CODE) {
        setPageError("你不是该板块版主或站点管理员");
        setCReportRows([]);
        setCReportTotal(0);
        return;
      }
      if (body.code !== API_SUCCESS_CODE || !body.data) {
        setPageError(apiErrorMessage(body));
        setCReportRows([]);
        setCReportTotal(0);
        return;
      }
      setCReportRows(Array.isArray(body.data.list) ? body.data.list : []);
      setCReportTotal(body.data.total ?? 0);
      setCPendingOpen(body.data.pending_open ?? 0);
      setCReportPage(body.data.page ?? cReportPage);
      setCSelectedReportIDs((prev) =>
        prev.filter((id) => (body.data.list ?? []).some((r) => r.id === id)),
      );
    } catch (e) {
      setPageError(e instanceof Error ? e.message : "评论举报加载失败");
      setCReportRows([]);
      setCReportTotal(0);
    } finally {
      setCReportLoading(false);
    }
  }, [board, cReportPage, pageSize, cReportStatus]);

  const loadDeleted = useCallback(async () => {
    if (!board) return;
    const token = getAccessToken();
    if (!token) return;
    setDelLoading(true);
    setPageError(null);
    try {
      const body = await apiListBoardDeletedPosts(token, board.id, delPage, pageSize);
      if (body.code === API_FORBIDDEN_CODE) {
        setPageError("你不是该板块版主或站点管理员");
        setDelRows([]);
        setDelTotal(0);
        return;
      }
      if (body.code !== API_SUCCESS_CODE || !body.data) {
        setPageError(apiErrorMessage(body));
        setDelRows([]);
        setDelTotal(0);
        return;
      }
      setDelRows(Array.isArray(body.data.list) ? body.data.list : []);
      setDelTotal(body.data.total ?? 0);
      setDelPage(body.data.page ?? delPage);
    } catch (e) {
      setPageError(e instanceof Error ? e.message : "已删帖列表加载失败");
      setDelRows([]);
      setDelTotal(0);
    } finally {
      setDelLoading(false);
    }
  }, [board, delPage, pageSize]);

  const loadAppeals = useCallback(async () => {
    if (!board) return;
    const token = getAccessToken();
    if (!token) return;
    setAppealLoading(true);
    setPageError(null);
    try {
      const body = await apiListBoardAppeals(
        token,
        board.id,
        appealPage,
        pageSize,
        appealStatus,
      );
      if (body.code === API_FORBIDDEN_CODE) {
        setPageError("你不是该板块版主或站点管理员");
        setAppealRows([]);
        setAppealTotal(0);
        return;
      }
      if (body.code !== API_SUCCESS_CODE || !body.data) {
        setPageError(apiErrorMessage(body));
        setAppealRows([]);
        setAppealTotal(0);
        return;
      }
      setAppealRows(Array.isArray(body.data.list) ? body.data.list : []);
      setAppealTotal(body.data.total ?? 0);
      setAppealPage(body.data.page ?? appealPage);
    } catch (e) {
      setPageError(e instanceof Error ? e.message : "申诉加载失败");
      setAppealRows([]);
      setAppealTotal(0);
    } finally {
      setAppealLoading(false);
    }
  }, [board, appealPage, pageSize, appealStatus]);

  const loadDashboard = useCallback(async () => {
    if (!board) return;
    const token = getAccessToken();
    if (!token) return;
    try {
      const body = await apiGetModerationDashboard(token, board.id);
      if (body.code === API_SUCCESS_CODE && body.data) {
        setDashboard(body.data);
      }
    } catch {
      // ignore dashboard refresh failures
    }
  }, [board]);

  const loadLogs = useCallback(async () => {
    if (!board) return;
    const token = getAccessToken();
    if (!token) return;
    setLogLoading(true);
    setPageError(null);
    try {
      const body = await apiListBoardModerationLogs(
        token,
        board.id,
        logPage,
        pageSize,
        logAction,
        logTargetType || undefined,
        logTargetIDInput.trim() ? Number.parseInt(logTargetIDInput, 10) : undefined,
      );
      if (body.code === API_FORBIDDEN_CODE) {
        setPageError("你不是该板块版主或站点管理员");
        setLogRows([]);
        setLogTotal(0);
        return;
      }
      if (body.code !== API_SUCCESS_CODE || !body.data) {
        setPageError(apiErrorMessage(body));
        setLogRows([]);
        setLogTotal(0);
        return;
      }
      setLogRows(Array.isArray(body.data.list) ? body.data.list : []);
      setLogTotal(body.data.total ?? 0);
      setLogPage(body.data.page ?? logPage);
    } catch (e) {
      setPageError(e instanceof Error ? e.message : "治理日志加载失败");
      setLogRows([]);
      setLogTotal(0);
    } finally {
      setLogLoading(false);
    }
  }, [board, logPage, pageSize, logAction, logTargetType, logTargetIDInput]);

  useEffect(() => {
    void loadBoard();
  }, [loadBoard]);

  useEffect(() => {
    void loadReports();
  }, [loadReports]);

  useEffect(() => {
    void loadCReports();
  }, [loadCReports]);

  useEffect(() => {
    void loadDeleted();
  }, [loadDeleted]);

  useEffect(() => {
    void loadAppeals();
  }, [loadAppeals]);

  useEffect(() => {
    void loadLogs();
  }, [loadLogs]);

  useEffect(() => {
    void loadDashboard();
  }, [loadDashboard]);

  async function onChangeReportStatus(
    reportID: number,
    next: PostReportStatus,
    currentNote: string,
  ) {
    if (!board) return;
    const token = getAccessToken();
    if (!token) return;
    setReportBusyID(reportID);
    setPageError(null);
    try {
      const note = window.prompt("可选：填写处理备注（留空可不改）", currentNote) ?? "";
      const res = await apiUpdateBoardReportStatus(token, board.id, reportID, {
        status: next,
        handler_note: note.trim(),
      });
      if (res.code !== API_SUCCESS_CODE) {
        if (res.code === API_POST_REPORT_NOT_EXIST_CODE) {
          setPageError("举报不存在，列表可能已变化");
          await loadReports();
          return;
        }
        if (res.code === API_FORBIDDEN_CODE) {
          setPageError("无权处理该举报");
          return;
        }
        setPageError(apiErrorMessage(res));
        return;
      }
      setFocusReportID(reportID);
      if (logAction !== "update_post_report_status") {
        setLogAction("update_post_report_status");
        setLogPage(1);
      }
      await Promise.all([loadReports(), loadLogs(), loadDashboard()]);
    } catch (e) {
      setPageError(e instanceof Error ? e.message : "操作失败");
    } finally {
      setReportBusyID(null);
    }
  }

  async function onBatchUpdateReports() {
    if (!board || selectedReportIDs.length === 0) return;
    const token = getAccessToken();
    if (!token) return;
    const note = window.prompt("可选：批量处理备注", "") ?? "";
    setBatchBusy(true);
    setPageError(null);
    try {
      const res = await apiBatchUpdateBoardReports(token, board.id, {
        report_ids: selectedReportIDs,
        status: batchStatus,
        handler_note: note.trim(),
      });
      if (res.code !== API_SUCCESS_CODE) {
        if (res.code === API_FORBIDDEN_CODE) {
          setPageError("无权执行批量处理");
          return;
        }
        if (res.code === API_POST_REPORT_NOT_EXIST_CODE) {
          setPageError("部分举报不存在或已不在当前板块");
          await loadReports();
          return;
        }
        setPageError(apiErrorMessage(res));
        return;
      }
      setSelectedReportIDs([]);
      setLogAction("update_post_report_status");
      setLogPage(1);
      await Promise.all([loadReports(), loadLogs(), loadDashboard()]);
    } catch (e) {
      setPageError(e instanceof Error ? e.message : "批量处理失败");
    } finally {
      setBatchBusy(false);
    }
  }

  async function onChangeCommentReportStatus(
    reportID: number,
    next: CommentReportStatus,
    currentNote: string,
  ) {
    if (!board) return;
    const token = getAccessToken();
    if (!token) return;
    setCReportBusyID(reportID);
    setPageError(null);
    try {
      const note = window.prompt("可选：填写处理备注（留空可不改）", currentNote) ?? "";
      const res = await apiUpdateBoardCommentReportStatus(token, board.id, reportID, {
        status: next,
        handler_note: note.trim(),
      });
      if (res.code !== API_SUCCESS_CODE) {
        if (res.code === API_COMMENT_REPORT_NOT_EXIST_CODE) {
          setPageError("评论举报不存在，列表可能已变化");
          await loadCReports();
          return;
        }
        if (res.code === API_FORBIDDEN_CODE) {
          setPageError("无权处理该评论举报");
          return;
        }
        setPageError(apiErrorMessage(res));
        return;
      }
      setFocusCommentReportID(reportID);
      if (logAction !== "update_comment_report_status") {
        setLogAction("update_comment_report_status");
        setLogPage(1);
      }
      await Promise.all([loadCReports(), loadLogs(), loadDashboard()]);
    } catch (e) {
      setPageError(e instanceof Error ? e.message : "操作失败");
    } finally {
      setCReportBusyID(null);
    }
  }

  async function onBatchUpdateCommentReports() {
    if (!board || cSelectedReportIDs.length === 0) return;
    const token = getAccessToken();
    if (!token) return;
    const note = window.prompt("可选：批量处理备注", "") ?? "";
    setCBatchBusy(true);
    setPageError(null);
    try {
      const res = await apiBatchUpdateBoardCommentReports(token, board.id, {
        report_ids: cSelectedReportIDs,
        status: cBatchStatus,
        handler_note: note.trim(),
      });
      if (res.code !== API_SUCCESS_CODE) {
        if (res.code === API_FORBIDDEN_CODE) {
          setPageError("无权执行批量处理");
          return;
        }
        if (res.code === API_COMMENT_REPORT_NOT_EXIST_CODE) {
          setPageError("部分评论举报不存在或已不在当前板块");
          await loadCReports();
          return;
        }
        setPageError(apiErrorMessage(res));
        return;
      }
      setCSelectedReportIDs([]);
      setLogAction("update_comment_report_status");
      setLogPage(1);
      await Promise.all([loadCReports(), loadLogs(), loadDashboard()]);
    } catch (e) {
      setPageError(e instanceof Error ? e.message : "批量处理失败");
    } finally {
      setCBatchBusy(false);
    }
  }

  async function onRestoreDeletedPost(postId: number) {
    const token = getAccessToken();
    if (!token) return;
    if (!window.confirm(`确定恢复帖子 #${postId}？`)) return;
    setRestoreBusyID(postId);
    setPageError(null);
    try {
      const res = await apiRestorePost(token, postId);
      if (res.code !== API_SUCCESS_CODE) {
        if (res.code === API_POST_NOT_SOFT_DELETED_CODE) {
          setPageError("该帖未处于软删状态，可能已被恢复");
          await loadDeleted();
          return;
        }
        if (res.code === API_FORBIDDEN_CODE) {
          setPageError("无权恢复该帖");
          return;
        }
        setPageError(apiErrorMessage(res));
        return;
      }
      await Promise.all([loadDeleted(), loadLogs(), loadDashboard()]);
    } catch (e) {
      setPageError(e instanceof Error ? e.message : "恢复失败");
    } finally {
      setRestoreBusyID(null);
    }
  }

  async function onHandleAppeal(
    appealID: number,
    next: "in_review" | "approved" | "rejected",
    postID: number,
  ) {
    if (!board) return;
    const token = getAccessToken();
    if (!token) return;
    setAppealBusyID(appealID);
    setPageError(null);
    try {
      const note = window.prompt("可选：填写给作者的回复", "") ?? "";
      let applyUpdate = false;
      if (next === "approved") {
        applyUpdate = window.confirm(
          "是否将帖子内容替换为作者的「修改稿」（标题+正文）？点「确定」应用；点「取消」仅解封不替换正文。",
        );
      }
      const res = await apiHandleBoardAppeal(token, board.id, appealID, {
        status: next,
        moderator_reply: note.trim(),
        apply_update: applyUpdate,
      });
      if (res.code !== API_SUCCESS_CODE) {
        if (res.code === API_POST_APPEAL_NOT_EXIST_CODE) {
          setPageError("申诉不存在，列表可能已变化");
          await loadAppeals();
          return;
        }
        if (res.code === API_FORBIDDEN_CODE) {
          setPageError("无权处理该申诉");
          return;
        }
        setPageError(apiErrorMessage(res));
        return;
      }
      setLogAction("handle_post_appeal");
      setLogTargetType("post");
      setLogTargetIDInput(String(postID));
      setLogPage(1);
      await Promise.all([loadAppeals(), loadLogs(), loadDashboard()]);
    } catch (e) {
      setPageError(e instanceof Error ? e.message : "申诉处理失败");
    } finally {
      setAppealBusyID(null);
    }
  }

  const reportTotalPages = Math.max(1, Math.ceil(reportTotal / pageSize));
  const cReportTotalPages = Math.max(1, Math.ceil(cReportTotal / pageSize));
  const delTotalPages = Math.max(1, Math.ceil(delTotal / pageSize));
  const appealTotalPages = Math.max(1, Math.ceil(appealTotal / pageSize));
  const logTotalPages = Math.max(1, Math.ceil(logTotal / pageSize));

  const highlightedLogIDs = useMemo(() => {
    const s = new Set<number>();
    if (focusReportID != null) {
      for (const row of logRows) {
        if (row.target_type === "post_report" && row.target_id === focusReportID) {
          s.add(row.id);
        }
      }
    }
    if (focusCommentReportID != null) {
      for (const row of logRows) {
        if (
          row.target_type === "comment_report" &&
          row.target_id === focusCommentReportID
        ) {
          s.add(row.id);
        }
      }
    }
    return s;
  }, [focusReportID, focusCommentReportID, logRows]);

  return (
    <div className="mx-auto max-w-6xl px-4 py-10">
      <button
        type="button"
        onClick={() => router.push(`/boards/${encodeURIComponent(slug)}`)}
        className="mb-6 text-sm text-zinc-500 underline"
      >
        ← 返回板块详情
      </button>

      <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
        版主治理工作台
      </h1>
      {board ? (
        <p className="mt-1 text-sm text-zinc-500">
          板块：{board.name} /{board.slug}
        </p>
      ) : null}
      <p className="mt-2 text-xs text-zinc-500">
        默认展示待处理举报与封帖申诉；支持批量处理举报、申诉处理（可应用修改稿并解封）、日志按目标筛选与联动。
      </p>
      {dashboard ? (
        <div className="mt-4 grid gap-2 sm:grid-cols-3">
          <div className="rounded border border-zinc-200 px-3 py-2 text-xs dark:border-zinc-800">
            待处理：{dashboard.pending_reports} · 处理中：{dashboard.in_review_reports}
          </div>
          <div className="rounded border border-zinc-200 px-3 py-2 text-xs dark:border-zinc-800">
            近 24h 新举报：{dashboard.reports_created_24h} · 近 24h 治理动作：{dashboard.logs_24h}
          </div>
          <div className="rounded border border-zinc-200 px-3 py-2 text-xs dark:border-zinc-800">
            近 7d 已解决：{dashboard.resolved_reports_7d} · 已驳回：{dashboard.rejected_reports_7d}
          </div>
        </div>
      ) : null}

      {loadingBoard ? <p className="mt-6 text-sm text-zinc-500">加载中…</p> : null}
      {pageError ? (
        <p className="mt-4 text-sm text-red-600 dark:text-red-400">{pageError}</p>
      ) : null}

      {!loadingBoard && board && !pageError ? (
        <>
        <div className="mt-6 grid gap-6 xl:grid-cols-3">
          <div className="space-y-6">
          <section className="rounded-xl border border-zinc-200 dark:border-zinc-800">
            <div className="border-b border-zinc-200 px-4 py-3 dark:border-zinc-800">
              <h2 className="text-sm font-medium text-zinc-800 dark:text-zinc-200">
                帖子举报
              </h2>
              <div className="mt-2 flex flex-wrap gap-2">
                {REPORT_STATUS_OPTIONS.map((opt) => (
                  <button
                    key={opt.id || "all"}
                    type="button"
                    onClick={() => {
                      setReportStatus(opt.id);
                      setReportPage(1);
                    }}
                    className={
                      reportStatus === opt.id
                        ? "rounded-lg border border-zinc-900 bg-zinc-900 px-3 py-1 text-xs text-white dark:border-zinc-100 dark:bg-zinc-100 dark:text-zinc-900"
                        : "rounded-lg border border-zinc-300 px-3 py-1 text-xs dark:border-zinc-600"
                    }
                  >
                    {opt.label}
                  </button>
                ))}
              </div>
              <p className="mt-2 text-xs text-amber-700 dark:text-amber-400">
                当前待处理：{pendingOpen}
              </p>
              <div className="mt-2 flex flex-wrap items-center gap-2">
                <select
                  value={batchStatus}
                  onChange={(e) => setBatchStatus(e.target.value as PostReportStatus)}
                  className="rounded border border-zinc-300 px-2 py-1 text-xs dark:border-zinc-600 dark:bg-zinc-900"
                >
                  <option value="in_review">批量设为处理中</option>
                  <option value="resolved">批量设为已解决</option>
                  <option value="rejected">批量设为已驳回</option>
                  <option value="open">批量设为待处理</option>
                </select>
                <button
                  type="button"
                  disabled={batchBusy || selectedReportIDs.length === 0}
                  onClick={() => void onBatchUpdateReports()}
                  className="rounded border border-zinc-300 px-3 py-1 text-xs disabled:opacity-40 dark:border-zinc-600"
                >
                  {batchBusy ? "处理中…" : `批量处理（${selectedReportIDs.length}）`}
                </button>
              </div>
            </div>
            <div className="max-h-[70vh] overflow-auto divide-y divide-zinc-200 dark:divide-zinc-800">
              {reportLoading ? (
                <p className="px-4 py-6 text-sm text-zinc-500">加载举报中…</p>
              ) : reportRows.length === 0 ? (
                <p className="px-4 py-6 text-sm text-zinc-500">暂无举报</p>
              ) : (
                reportRows.map((r) => (
                  <article key={r.id} className="px-4 py-4">
                    <div className="mb-2">
                      <label className="inline-flex items-center gap-2 text-xs text-zinc-500">
                        <input
                          type="checkbox"
                          checked={selectedReportIDs.includes(r.id)}
                          onChange={(e) =>
                            setSelectedReportIDs((prev) =>
                              e.target.checked
                                ? Array.from(new Set([...prev, r.id]))
                                : prev.filter((id) => id !== r.id),
                            )
                          }
                        />
                        选中用于批量处理
                      </label>
                    </div>
                    <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                      #{r.id} ·{" "}
                      <Link href={`/posts/${r.post_id}`} className="underline">
                        {r.post_title || `帖子 ${r.post_id}`}
                      </Link>
                    </p>
                    <p className="mt-1 text-xs text-zinc-500">
                      举报人：{r.reporter_username || r.reporter_id} · 状态：
                      {REPORT_STATUS_LABEL[r.status]}
                    </p>
                    <p className="mt-2 text-sm text-zinc-800 dark:text-zinc-200">
                      原因：{r.reason}
                    </p>
                    {r.detail ? (
                      <p className="mt-1 whitespace-pre-wrap text-sm text-zinc-600 dark:text-zinc-300">
                        {r.detail}
                      </p>
                    ) : null}
                    <div className="mt-3 flex items-center gap-2">
                      <select
                        value={r.status}
                        disabled={reportBusyID === r.id}
                        onChange={(e) =>
                          void onChangeReportStatus(
                            r.id,
                            e.target.value as PostReportStatus,
                            r.handler_note ?? "",
                          )
                        }
                        className="rounded border border-zinc-300 px-2 py-1 text-xs dark:border-zinc-600 dark:bg-zinc-900"
                      >
                        <option value="open">待处理</option>
                        <option value="in_review">处理中</option>
                        <option value="resolved">已解决</option>
                        <option value="rejected">已驳回</option>
                      </select>
                      <button
                        type="button"
                        onClick={() => {
                          setFocusReportID(r.id);
                          setLogAction("update_post_report_status");
                          setLogPage(1);
                        }}
                        className="rounded border border-zinc-300 px-2 py-1 text-xs dark:border-zinc-600"
                      >
                        查看关联日志
                      </button>
                    </div>
                    <p className="mt-2 text-xs text-zinc-500">
                      处理人：{r.handler_username || (r.handler_id ?? "-")} · 更新于{" "}
                      {new Date(r.update_time).toLocaleString("zh-CN")}
                    </p>
                    {r.handler_note ? (
                      <p className="mt-1 whitespace-pre-wrap text-xs text-zinc-500">
                        处理备注：{r.handler_note}
                      </p>
                    ) : null}
                  </article>
                ))
              )}
            </div>
            {!reportLoading && reportRows.length > 0 && reportTotal > pageSize ? (
              <div className="flex items-center justify-between border-t border-zinc-200 px-4 py-3 text-sm dark:border-zinc-800">
                <span className="text-zinc-500">
                  第 {reportPage} / {reportTotalPages} 页 · 共 {reportTotal} 条
                </span>
                <div className="flex gap-2">
                  <button
                    type="button"
                    disabled={reportPage <= 1}
                    onClick={() => setReportPage((p) => Math.max(1, p - 1))}
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    上一页
                  </button>
                  <button
                    type="button"
                    disabled={reportPage >= reportTotalPages}
                    onClick={() =>
                      setReportPage((p) => Math.min(reportTotalPages, p + 1))
                    }
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    下一页
                  </button>
                </div>
              </div>
            ) : null}
          </section>

          <section className="rounded-xl border border-zinc-200 dark:border-zinc-800">
            <div className="border-b border-zinc-200 px-4 py-3 dark:border-zinc-800">
              <h2 className="text-sm font-medium text-zinc-800 dark:text-zinc-200">
                评论举报
              </h2>
              <div className="mt-2 flex flex-wrap gap-2">
                {COMMENT_REPORT_STATUS_OPTIONS.map((opt) => (
                  <button
                    key={`c-${opt.id || "all"}`}
                    type="button"
                    onClick={() => {
                      setCReportStatus(opt.id);
                      setCReportPage(1);
                    }}
                    className={
                      cReportStatus === opt.id
                        ? "rounded-lg border border-zinc-900 bg-zinc-900 px-3 py-1 text-xs text-white dark:border-zinc-100 dark:bg-zinc-100 dark:text-zinc-900"
                        : "rounded-lg border border-zinc-300 px-3 py-1 text-xs dark:border-zinc-600"
                    }
                  >
                    {opt.label}
                  </button>
                ))}
              </div>
              <p className="mt-2 text-xs text-amber-700 dark:text-amber-400">
                当前待处理：{cPendingOpen}
              </p>
              <div className="mt-2 flex flex-wrap items-center gap-2">
                <select
                  value={cBatchStatus}
                  onChange={(e) => setCBatchStatus(e.target.value as CommentReportStatus)}
                  className="rounded border border-zinc-300 px-2 py-1 text-xs dark:border-zinc-600 dark:bg-zinc-900"
                >
                  <option value="in_review">批量设为处理中</option>
                  <option value="resolved">批量设为已解决</option>
                  <option value="rejected">批量设为已驳回</option>
                  <option value="open">批量设为待处理</option>
                </select>
                <button
                  type="button"
                  disabled={cBatchBusy || cSelectedReportIDs.length === 0}
                  onClick={() => void onBatchUpdateCommentReports()}
                  className="rounded border border-zinc-300 px-3 py-1 text-xs disabled:opacity-40 dark:border-zinc-600"
                >
                  {cBatchBusy ? "处理中…" : `批量处理评论举报（${cSelectedReportIDs.length}）`}
                </button>
              </div>
            </div>
            <div className="max-h-[70vh] overflow-auto divide-y divide-zinc-200 dark:divide-zinc-800">
              {cReportLoading ? (
                <p className="px-4 py-6 text-sm text-zinc-500">加载评论举报中…</p>
              ) : cReportRows.length === 0 ? (
                <p className="px-4 py-6 text-sm text-zinc-500">暂无评论举报</p>
              ) : (
                cReportRows.map((r) => (
                  <article key={r.id} className="px-4 py-4">
                    <div className="mb-2">
                      <label className="inline-flex items-center gap-2 text-xs text-zinc-500">
                        <input
                          type="checkbox"
                          checked={cSelectedReportIDs.includes(r.id)}
                          onChange={(e) =>
                            setCSelectedReportIDs((prev) =>
                              e.target.checked
                                ? Array.from(new Set([...prev, r.id]))
                                : prev.filter((id) => id !== r.id),
                            )
                          }
                        />
                        选中用于批量处理
                      </label>
                    </div>
                    <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                      #{r.id} · 评论 #{r.comment_id} ·{" "}
                      <Link href={`/posts/${r.post_id}`} className="underline">
                        {r.post_title || `帖子 ${r.post_id}`}
                      </Link>
                    </p>
                    <p className="mt-1 line-clamp-3 text-xs text-zinc-600 dark:text-zinc-300">
                      摘录：{r.comment_snippet || "（无正文）"}
                    </p>
                    <p className="mt-1 text-xs text-zinc-500">
                      举报人：{r.reporter_username || r.reporter_id} · 状态：
                      {COMMENT_REPORT_STATUS_LABEL[r.status]}
                    </p>
                    <p className="mt-2 text-sm text-zinc-800 dark:text-zinc-200">
                      原因：{r.reason}
                    </p>
                    {r.detail ? (
                      <p className="mt-1 whitespace-pre-wrap text-sm text-zinc-600 dark:text-zinc-300">
                        {r.detail}
                      </p>
                    ) : null}
                    <div className="mt-3 flex items-center gap-2">
                      <select
                        value={r.status}
                        disabled={cReportBusyID === r.id}
                        onChange={(e) =>
                          void onChangeCommentReportStatus(
                            r.id,
                            e.target.value as CommentReportStatus,
                            r.handler_note ?? "",
                          )
                        }
                        className="rounded border border-zinc-300 px-2 py-1 text-xs dark:border-zinc-600 dark:bg-zinc-900"
                      >
                        <option value="open">待处理</option>
                        <option value="in_review">处理中</option>
                        <option value="resolved">已解决</option>
                        <option value="rejected">已驳回</option>
                      </select>
                      <button
                        type="button"
                        onClick={() => {
                          setFocusCommentReportID(r.id);
                          setLogAction("update_comment_report_status");
                          setLogPage(1);
                        }}
                        className="rounded border border-zinc-300 px-2 py-1 text-xs dark:border-zinc-600"
                      >
                        查看关联日志
                      </button>
                    </div>
                    <p className="mt-2 text-xs text-zinc-500">
                      处理人：{r.handler_username || (r.handler_id ?? "-")} · 更新于{" "}
                      {new Date(r.update_time).toLocaleString("zh-CN")}
                    </p>
                    {r.handler_note ? (
                      <p className="mt-1 whitespace-pre-wrap text-xs text-zinc-500">
                        处理备注：{r.handler_note}
                      </p>
                    ) : null}
                  </article>
                ))
              )}
            </div>
            {!cReportLoading && cReportRows.length > 0 && cReportTotal > pageSize ? (
              <div className="flex items-center justify-between border-t border-zinc-200 px-4 py-3 text-sm dark:border-zinc-800">
                <span className="text-zinc-500">
                  第 {cReportPage} / {cReportTotalPages} 页 · 共 {cReportTotal} 条
                </span>
                <div className="flex gap-2">
                  <button
                    type="button"
                    disabled={cReportPage <= 1}
                    onClick={() => setCReportPage((p) => Math.max(1, p - 1))}
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    上一页
                  </button>
                  <button
                    type="button"
                    disabled={cReportPage >= cReportTotalPages}
                    onClick={() =>
                      setCReportPage((p) => Math.min(cReportTotalPages, p + 1))
                    }
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    下一页
                  </button>
                </div>
              </div>
            ) : null}
          </section>
          </div>

          <section className="rounded-xl border border-zinc-200 dark:border-zinc-800">
            <div className="border-b border-zinc-200 px-4 py-3 dark:border-zinc-800">
              <h2 className="text-sm font-medium text-zinc-800 dark:text-zinc-200">
                封帖申诉
              </h2>
              <div className="mt-2 flex flex-wrap gap-2">
                {APPEAL_STATUS_OPTIONS.map((opt) => (
                  <button
                    key={opt.id || "all"}
                    type="button"
                    onClick={() => {
                      setAppealStatus(opt.id);
                      setAppealPage(1);
                    }}
                    className={
                      appealStatus === opt.id
                        ? "rounded-lg border border-zinc-900 bg-zinc-900 px-3 py-1 text-xs text-white dark:border-zinc-100 dark:bg-zinc-100 dark:text-zinc-900"
                        : "rounded-lg border border-zinc-300 px-3 py-1 text-xs dark:border-zinc-600"
                    }
                  >
                    {opt.label}
                  </button>
                ))}
              </div>
              <p className="mt-2 text-xs text-zinc-500">
                通过申诉时可选择是否把帖子替换为作者的修改稿，并自动解封。
              </p>
            </div>
            <div className="max-h-[70vh] overflow-auto divide-y divide-zinc-200 dark:divide-zinc-800">
              {appealLoading ? (
                <p className="px-4 py-6 text-sm text-zinc-500">加载申诉中…</p>
              ) : appealRows.length === 0 ? (
                <p className="px-4 py-6 text-sm text-zinc-500">暂无申诉</p>
              ) : (
                appealRows.map((a) => (
                  <article key={a.id} className="px-4 py-4">
                    <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                      申诉 #{a.id} ·{" "}
                      <Link href={`/posts/${a.post_id}`} className="underline">
                        {a.post_title || `帖子 ${a.post_id}`}
                      </Link>
                    </p>
                    <p className="mt-1 text-xs text-zinc-500">
                      作者：{a.author_username || a.author_id} · 状态：
                      {APPEAL_STATUS_LABEL[a.status]}
                    </p>
                    <p className="mt-2 whitespace-pre-wrap text-sm text-zinc-800 dark:text-zinc-200">
                      申诉说明：{a.reason}
                    </p>
                    {a.user_reply ? (
                      <p className="mt-2 whitespace-pre-wrap text-sm text-zinc-700 dark:text-zinc-300">
                        作者补充：{a.user_reply}
                      </p>
                    ) : null}
                    <div className="mt-3 rounded border border-zinc-200 bg-zinc-50 p-3 text-xs dark:border-zinc-800 dark:bg-zinc-900">
                      <p className="font-medium text-zinc-800 dark:text-zinc-200">修改稿预览</p>
                      <p className="mt-1 text-sm text-zinc-900 dark:text-zinc-100">{a.requested_title}</p>
                      <p className="mt-2 whitespace-pre-wrap text-sm text-zinc-700 dark:text-zinc-300">
                        {a.requested_content}
                      </p>
                    </div>
                    {a.moderator_reply ? (
                      <p className="mt-2 whitespace-pre-wrap text-xs text-zinc-500">
                        版主回复：{a.moderator_reply}
                      </p>
                    ) : null}
                    <div className="mt-3 flex flex-wrap items-center gap-2">
                      <select
                        value={a.status}
                        disabled={appealBusyID === a.id}
                        onChange={(e) =>
                          void onHandleAppeal(
                            a.id,
                            e.target.value as "in_review" | "approved" | "rejected",
                            a.post_id,
                          )
                        }
                        className="rounded border border-zinc-300 px-2 py-1 text-xs dark:border-zinc-600 dark:bg-zinc-900"
                      >
                        <option value="open">待处理</option>
                        <option value="in_review">处理中</option>
                        <option value="approved">通过（可解封）</option>
                        <option value="rejected">驳回</option>
                      </select>
                      <button
                        type="button"
                        onClick={() => {
                          setLogAction("handle_post_appeal");
                          setLogTargetType("post");
                          setLogTargetIDInput(String(a.post_id));
                          setLogPage(1);
                        }}
                        className="rounded border border-zinc-300 px-2 py-1 text-xs dark:border-zinc-600"
                      >
                        查看关联日志
                      </button>
                    </div>
                    <p className="mt-2 text-xs text-zinc-500">
                      更新于 {new Date(a.update_time).toLocaleString("zh-CN")}
                    </p>
                  </article>
                ))
              )}
            </div>
            {!appealLoading && appealRows.length > 0 && appealTotal > pageSize ? (
              <div className="flex items-center justify-between border-t border-zinc-200 px-4 py-3 text-sm dark:border-zinc-800">
                <span className="text-zinc-500">
                  第 {appealPage} / {appealTotalPages} 页 · 共 {appealTotal} 条
                </span>
                <div className="flex gap-2">
                  <button
                    type="button"
                    disabled={appealPage <= 1}
                    onClick={() => setAppealPage((p) => Math.max(1, p - 1))}
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    上一页
                  </button>
                  <button
                    type="button"
                    disabled={appealPage >= appealTotalPages}
                    onClick={() =>
                      setAppealPage((p) => Math.min(appealTotalPages, p + 1))
                    }
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    下一页
                  </button>
                </div>
              </div>
            ) : null}
          </section>

          <section className="rounded-xl border border-zinc-200 dark:border-zinc-800">
            <div className="border-b border-zinc-200 px-4 py-3 dark:border-zinc-800">
              <h2 className="text-sm font-medium text-zinc-800 dark:text-zinc-200">
                治理日志
              </h2>
              <div className="mt-2 flex flex-wrap gap-2">
                {LOG_ACTION_OPTIONS.map((opt) => (
                  <button
                    key={opt.id || "all"}
                    type="button"
                    onClick={() => {
                      setLogAction(opt.id);
                      setLogPage(1);
                    }}
                    className={
                      logAction === opt.id
                        ? "rounded-lg border border-zinc-900 bg-zinc-900 px-3 py-1 text-xs text-white dark:border-zinc-100 dark:bg-zinc-100 dark:text-zinc-900"
                        : "rounded-lg border border-zinc-300 px-3 py-1 text-xs dark:border-zinc-600"
                    }
                  >
                    {opt.label}
                  </button>
                ))}
              </div>
              <div className="mt-2 flex flex-wrap items-center gap-2">
                <select
                  value={logTargetType}
                  onChange={(e) =>
                    setLogTargetType(
                      e.target.value as
                        | ""
                        | "post"
                        | "post_report"
                        | "comment_report"
                        | "board_moderator",
                    )
                  }
                  className="rounded border border-zinc-300 px-2 py-1 text-xs dark:border-zinc-600 dark:bg-zinc-900"
                >
                  <option value="">全部目标类型</option>
                  <option value="post">post</option>
                  <option value="post_report">post_report</option>
                  <option value="comment_report">comment_report</option>
                  <option value="board_moderator">board_moderator</option>
                </select>
                <input
                  value={logTargetIDInput}
                  onChange={(e) => setLogTargetIDInput(e.target.value)}
                  placeholder="target_id（可选）"
                  className="w-40 rounded border border-zinc-300 px-2 py-1 text-xs dark:border-zinc-600 dark:bg-zinc-900"
                />
              </div>
            </div>
            <div className="max-h-[70vh] overflow-auto divide-y divide-zinc-200 dark:divide-zinc-800">
              {logLoading ? (
                <p className="px-4 py-6 text-sm text-zinc-500">加载日志中…</p>
              ) : logRows.length === 0 ? (
                <p className="px-4 py-6 text-sm text-zinc-500">暂无日志</p>
              ) : (
                logRows.map((r) => (
                  <article
                    key={r.id}
                    className={
                      highlightedLogIDs.has(r.id)
                        ? "bg-amber-50/70 px-4 py-4 dark:bg-amber-950/30"
                        : "px-4 py-4"
                    }
                  >
                    <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                      {LOG_ACTION_LABEL[r.action]} · 目标 {r.target_type}:{r.target_id}
                    </p>
                    <p className="mt-1 text-xs text-zinc-500">
                      操作者：{r.operator_username || r.operator_id} ·{" "}
                      {new Date(r.create_time).toLocaleString("zh-CN")}
                    </p>
                    {r.description ? (
                      <p className="mt-2 whitespace-pre-wrap text-sm text-zinc-700 dark:text-zinc-300">
                        {r.description}
                      </p>
                    ) : null}
                  </article>
                ))
              )}
            </div>
            {!logLoading && logRows.length > 0 && logTotal > pageSize ? (
              <div className="flex items-center justify-between border-t border-zinc-200 px-4 py-3 text-sm dark:border-zinc-800">
                <span className="text-zinc-500">
                  第 {logPage} / {logTotalPages} 页 · 共 {logTotal} 条
                </span>
                <div className="flex gap-2">
                  <button
                    type="button"
                    disabled={logPage <= 1}
                    onClick={() => setLogPage((p) => Math.max(1, p - 1))}
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    上一页
                  </button>
                  <button
                    type="button"
                    disabled={logPage >= logTotalPages}
                    onClick={() => setLogPage((p) => Math.min(logTotalPages, p + 1))}
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    下一页
                  </button>
                </div>
              </div>
            ) : null}
          </section>
        </div>

        <section className="mt-6 rounded-xl border border-zinc-200 dark:border-zinc-800">
          <div className="border-b border-zinc-200 px-4 py-3 dark:border-zinc-800">
            <h2 className="text-sm font-medium text-zinc-800 dark:text-zinc-200">
              本板已软删帖子（可恢复）
            </h2>
          </div>
          <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
            {delLoading ? (
              <p className="px-4 py-6 text-sm text-zinc-500">加载中…</p>
            ) : delRows.length === 0 ? (
              <p className="px-4 py-6 text-sm text-zinc-500">暂无已软删帖子</p>
            ) : (
              delRows.map((row) => (
                <div
                  key={row.id}
                  className="flex flex-wrap items-center justify-between gap-3 px-4 py-3"
                >
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                      #{row.id} · {row.title || "（无标题）"}
                    </p>
                    <p className="mt-1 text-xs text-zinc-500">
                      删除于{" "}
                      {new Date(row.deleted_at).toLocaleString("zh-CN")}
                      {row.author_id != null ? ` · 作者 ID ${row.author_id}` : ""}
                    </p>
                  </div>
                  <button
                    type="button"
                    disabled={restoreBusyID === row.id}
                    onClick={() => void onRestoreDeletedPost(row.id)}
                    className="shrink-0 rounded border border-emerald-600 px-3 py-1 text-xs text-emerald-800 disabled:opacity-40 dark:border-emerald-700 dark:text-emerald-200"
                  >
                    {restoreBusyID === row.id ? "恢复中…" : "恢复"}
                  </button>
                </div>
              ))
            )}
          </div>
          {!delLoading && delRows.length > 0 && delTotal > pageSize ? (
            <div className="flex items-center justify-between border-t border-zinc-200 px-4 py-3 text-sm dark:border-zinc-800">
              <span className="text-zinc-500">
                第 {delPage} / {delTotalPages} 页 · 共 {delTotal} 条
              </span>
              <div className="flex gap-2">
                <button
                  type="button"
                  disabled={delPage <= 1}
                  onClick={() => setDelPage((p) => Math.max(1, p - 1))}
                  className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                >
                  上一页
                </button>
                <button
                  type="button"
                  disabled={delPage >= delTotalPages}
                  onClick={() => setDelPage((p) => Math.min(delTotalPages, p + 1))}
                  className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                >
                  下一页
                </button>
              </div>
            </div>
          ) : null}
        </section>
        </>
      ) : null}

      <p className="mt-6 text-sm text-zinc-500">
        兼容入口：你仍可使用
        {" "}
        <Link href={`/boards/${encodeURIComponent(slug)}/reports`} className="underline">
          旧举报页
        </Link>
        {" "}
        与
        {" "}
        <Link href={`/boards/${encodeURIComponent(slug)}/mod-logs`} className="underline">
          旧日志页
        </Link>
        。
      </p>
    </div>
  );
}
